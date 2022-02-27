package dashboard

import (
	"errors"
	"fmt"
	"net/http"

	dashv1alpha1 "github.com/cosmo-workspace/cosmo/api/openapi/dashboard/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/google/uuid"
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

func NormalResponse(code int, body interface{}) (dashv1alpha1.ImplResponse, error) {
	return dashv1alpha1.Response(code, body), nil
}

func ErrorResponse(code int, message string) (dashv1alpha1.ImplResponse, error) {
	return dashv1alpha1.Response(code, nil), errors.New(message)
}
