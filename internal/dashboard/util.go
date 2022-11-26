package dashboard

import (
	"errors"
	"fmt"
	"net/http"

	connect_go "github.com/bufbuild/connect-go"
	"github.com/google/uuid"

	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
)

type StoreStatusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *StoreStatusResponseWriter) StatusCode() int {
	return w.statusCode
}

func (w *StoreStatusResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *StoreStatusResponseWriter) StatusString() string {
	return http.StatusText(w.statusCode)
}

type HTTPRequestLogger struct {
	*clog.Logger
}

func NewHTTPRequestLogger(logr *clog.Logger) HTTPRequestLogger {
	return HTTPRequestLogger{logr}
}

func (l HTTPRequestLogger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := l.Logger.WithValues("reqID", uuid.New().String())

		ctx := clog.IntoContext(r.Context(), log)
		r = r.WithContext(ctx)

		rw := &StoreStatusResponseWriter{ResponseWriter: w}
		next.ServeHTTP(rw, r)

		log.WithName("access").Info(fmt.Sprintf("%d %s %s %s", rw.StatusCode(), rw.StatusString(), r.Method, r.URL),
			"method", r.Method, "path", r.URL, "statusCode", rw.StatusCode(), "host", r.Host, "X-Forwarded-For", r.Header.Get("X-Forwarded-For"), "user-agent", r.UserAgent())
	})
}

func ErrResponse(log *clog.Logger, err error) error {

	if errors.Is(err, &connect_go.Error{}) {
		// pass
	} else if errors.Is(err, kosmo.ErrNotFound) {
		err = connect_go.NewError(connect_go.CodeNotFound, err)

	} else if errors.Is(err, kosmo.ErrIsAlreadyExists) {
		err = connect_go.NewError(connect_go.CodeAlreadyExists, err)

	} else if errors.Is(err, kosmo.ErrBadRequest) {
		err = connect_go.NewError(connect_go.CodeInvalidArgument, err)

	} else if errors.Is(err, kosmo.ErrForbidden) {
		err = connect_go.NewError(connect_go.CodePermissionDenied, err)

	} else if errors.Is(err, kosmo.ErrUnauthorized) {
		err = connect_go.NewError(connect_go.CodeUnauthenticated, err)

	} else if errors.Is(err, kosmo.ErrServiceUnavailable) {
		err = connect_go.NewError(connect_go.CodeUnavailable, err)

	} else if errors.Is(err, kosmo.ErrInternalServerError) {
		err = connect_go.NewError(connect_go.CodeInternal, err)

	} else {
		err = connect_go.NewError(connect_go.CodeInternal, err)

	}
	log.WithCaller().Info(err.Error())
	return err
}
