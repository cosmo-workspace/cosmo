package dashboard

import (
	"context"

	connect_go "github.com/bufbuild/connect-go"

	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
)

type validator interface {
	Validate() error
}

func (s *Server) validatorInterceptor() connect_go.UnaryInterceptorFunc {
	interceptor := func(next connect_go.UnaryFunc) connect_go.UnaryFunc {
		return connect_go.UnaryFunc(func(ctx context.Context, req connect_go.AnyRequest) (connect_go.AnyResponse, error) {
			log := clog.FromContext(ctx).WithName("validator")

			if v, ok := req.Any().(validator); ok {
				if err := v.Validate(); err != nil {
					return nil, ErrResponse(log, kosmo.NewBadRequestError(err.Error(), nil))
				}
			}
			return next(ctx, req)
		})
	}
	return connect_go.UnaryInterceptorFunc(interceptor)
}
