package dashboard

import (
	"errors"
	"fmt"
	"net/http"

	connect_go "github.com/bufbuild/connect-go"
	"github.com/google/uuid"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/cosmo-workspace/cosmo/pkg/clog"
)

type StoreStatusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *StoreStatusResponseWriter) StatusCode() int {
	return w.statusCode
}

func (w *StoreStatusResponseWriter) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
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
	} else if apierrs.IsNotFound(err) {
		err = connect_go.NewError(connect_go.CodeNotFound, err)

	} else if apierrs.IsAlreadyExists(err) {
		err = connect_go.NewError(connect_go.CodeAlreadyExists, err)

	} else if apierrs.IsBadRequest(err) {
		err = connect_go.NewError(connect_go.CodeInvalidArgument, err)

	} else if apierrs.IsForbidden(err) {
		err = connect_go.NewError(connect_go.CodePermissionDenied, err)

	} else if apierrs.IsUnauthorized(err) {
		err = connect_go.NewError(connect_go.CodeUnauthenticated, err)

	} else if apierrs.IsServiceUnavailable(err) {
		err = connect_go.NewError(connect_go.CodeUnavailable, err)

	} else if apierrs.IsInternalError(err) {
		err = connect_go.NewError(connect_go.CodeInternal, err)

	} else {
		err = connect_go.NewError(connect_go.CodeInternal, err)

	}
	log.WithCaller().Info(err.Error())
	return err
}

func NewForbidden(err error) error {
	return apierrs.NewForbidden(schema.GroupResource{}, "", err)
}
