package logto

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/Sokol111/ecommerce-tenant-service/internal/application/tenant"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type logtoClient struct {
	baseURL     string
	tokenSource oauth2.TokenSource
	httpClient  *http.Client
	log         *zap.Logger

	roleCache   map[string]string
	roleCacheMu sync.RWMutex
}

func newLogtoClient(cfg Config, tokenSource oauth2.TokenSource, log *zap.Logger) (tenant.IdentityProvider, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("logto base-url is required")
	}

	c := &logtoClient{
		baseURL:     cfg.BaseURL,
		tokenSource: tokenSource,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		log:         log.Named("logto"),
		roleCache:   make(map[string]string),
	}

	return c, nil
}

func (c *logtoClient) CreateUser(ctx context.Context, params tenant.CreateUserParams) (string, error) {
	body := map[string]any{
		"primaryEmail": params.Email,
		"password":     params.Password,
		"name":         params.FirstName + " " + params.LastName,
	}

	var result struct {
		ID string `json:"id"`
	}

	statusCode, err := c.doRequest(ctx, http.MethodPost, "/api/users", body, &result)
	if statusCode == http.StatusConflict {
		return c.handleUserConflict(ctx, params)
	}
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	c.log.Debug("user created in Logto", zap.String("userID", result.ID))
	return result.ID, nil
}

// handleUserConflict handles 409 from CreateUser: if existing user is an orphan
// (no tenant assigned), delete and recreate. Otherwise return ErrUserAlreadyExists.
func (c *logtoClient) handleUserConflict(ctx context.Context, params tenant.CreateUserParams) (string, error) {
	existingID, hasTenant, err := c.findUserByEmail(ctx, params.Email)
	if err != nil {
		return "", fmt.Errorf("failed to look up existing user: %w", err)
	}

	if hasTenant {
		return "", tenant.ErrUserAlreadyExists
	}

	// Orphaned user — delete and recreate
	c.log.Info("deleting orphaned Logto user before recreate", zap.String("userID", existingID))
	if err := c.DeleteUser(ctx, existingID); err != nil {
		return "", fmt.Errorf("failed to delete orphaned user: %w", err)
	}

	body := map[string]any{
		"primaryEmail": params.Email,
		"password":     params.Password,
		"name":         params.FirstName + " " + params.LastName,
	}

	var result struct {
		ID string `json:"id"`
	}

	if _, err := c.doRequest(ctx, http.MethodPost, "/api/users", body, &result); err != nil {
		return "", fmt.Errorf("failed to recreate user: %w", err)
	}

	c.log.Debug("user recreated in Logto", zap.String("userID", result.ID))
	return result.ID, nil
}

func (c *logtoClient) findUserByEmail(ctx context.Context, email string) (id string, hasTenant bool, err error) {
	var users []struct {
		ID         string `json:"id"`
		CustomData struct {
			Tenant string `json:"tenant"`
		} `json:"customData"`
	}

	_, err = c.doRequest(ctx, http.MethodGet, "/api/users?search.primaryEmail="+email, nil, &users)
	if err != nil {
		return "", false, err
	}

	if len(users) == 0 {
		return "", false, fmt.Errorf("user with email %q not found despite 409", email)
	}

	return users[0].ID, users[0].CustomData.Tenant != "", nil
}

func (c *logtoClient) SetUserTenant(ctx context.Context, userID string, tenantSlug string) error {
	body := map[string]any{
		"customData": map[string]any{
			"tenant": tenantSlug,
		},
	}

	_, err := c.doRequest(ctx, http.MethodPatch, "/api/users/"+userID, body, nil)
	if err != nil {
		return fmt.Errorf("failed to set user tenant: %w", err)
	}

	c.log.Debug("tenant set on user", zap.String("userID", userID), zap.String("tenant", tenantSlug))
	return nil
}

func (c *logtoClient) AssignRole(ctx context.Context, userID string, roleName string) error {
	roleID, err := c.resolveRoleID(ctx, roleName)
	if err != nil {
		return fmt.Errorf("failed to resolve role %q: %w", roleName, err)
	}

	body := map[string]any{
		"roleIds": []string{roleID},
	}

	statusCode, err := c.doRequest(ctx, http.MethodPost, "/api/users/"+userID+"/roles", body, nil)
	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}
	// 409 means role already assigned — idempotent
	if statusCode == http.StatusConflict {
		c.log.Debug("role already assigned", zap.String("userID", userID), zap.String("role", roleName))
		return nil
	}

	c.log.Debug("role assigned", zap.String("userID", userID), zap.String("role", roleName))
	return nil
}

func (c *logtoClient) DeleteUser(ctx context.Context, userID string) error {
	statusCode, err := c.doRequest(ctx, http.MethodDelete, "/api/users/"+userID, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	// 404 means already deleted — idempotent
	if statusCode == http.StatusNotFound {
		c.log.Debug("user already deleted", zap.String("userID", userID))
		return nil
	}

	c.log.Debug("user deleted from Logto", zap.String("userID", userID))
	return nil
}

func (c *logtoClient) resolveRoleID(ctx context.Context, roleName string) (string, error) {
	c.roleCacheMu.RLock()
	if id, ok := c.roleCache[roleName]; ok {
		c.roleCacheMu.RUnlock()
		return id, nil
	}
	c.roleCacheMu.RUnlock()

	c.roleCacheMu.Lock()
	defer c.roleCacheMu.Unlock()

	// Double-check after acquiring write lock
	if id, ok := c.roleCache[roleName]; ok {
		return id, nil
	}

	var roles []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	_, err := c.doRequest(ctx, http.MethodGet, "/api/roles?search="+roleName, nil, &roles)
	if err != nil {
		return "", err
	}

	for _, r := range roles {
		if r.Name == roleName {
			c.roleCache[roleName] = r.ID
			return r.ID, nil
		}
	}

	return "", fmt.Errorf("role %q not found in Logto", roleName)
}

func (c *logtoClient) doRequest(ctx context.Context, method, path string, body any, result any) (int, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return 0, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	token, err := c.tokenSource.Token()
	if err != nil {
		return 0, fmt.Errorf("failed to get access token: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // response body close error is not actionable

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 && resp.StatusCode != http.StatusConflict && resp.StatusCode != http.StatusNotFound {
		return resp.StatusCode, fmt.Errorf("logto API error: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	if result != nil && len(respBody) > 0 && resp.StatusCode < 400 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return resp.StatusCode, fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return resp.StatusCode, nil
}
