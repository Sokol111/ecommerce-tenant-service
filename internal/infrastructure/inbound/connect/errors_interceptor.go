package connect

import (
	"context"

	"connectrpc.com/connect"
	interceptor "github.com/Sokol111/ecommerce-commons/pkg/http/connect/interceptor"
	"go.uber.org/fx"
)

// ErrorMappingModule provides an interceptor that maps domain errors to connect.Error codes.
// Recommended priority: 25 (after logger=20, before handler, so logger sees mapped codes).
func ErrorMappingModule(priority int) fx.Option {
	return fx.Provide(
		fx.Annotate(
			func() interceptor.Interceptor {
				return interceptor.Interceptor{
					Priority: priority,
					Handler:  connect.UnaryInterceptorFunc(errorMappingUnaryInterceptor),
				}
			},
			fx.ResultTags(`group:"connect_interceptor"`),
		),
	)
}

func errorMappingUnaryInterceptor(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		resp, err := next(ctx, req)
		if err != nil {
			return nil, mapConnectError(err)
		}
		return resp, nil
	}
}
