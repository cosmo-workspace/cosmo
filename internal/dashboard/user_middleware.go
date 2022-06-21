package dashboard

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/pointer"

	dashv1alpha1 "github.com/cosmo-workspace/cosmo/api/openapi/dashboard/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
)

type ctxKeyUser struct{}

func newContextWithUser(ctx context.Context, user *wsv1alpha1.User) context.Context {
	return context.WithValue(ctx, ctxKeyUser{}, user)
}

func userFromContext(ctx context.Context) *wsv1alpha1.User {
	user, ok := ctx.Value(ctxKeyUser{}).(*wsv1alpha1.User)
	if ok && user != nil {
		return user.DeepCopy()
	}
	return nil
}

func (s *Server) preFetchUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := clog.FromContext(ctx).WithName("preFetchUser")

		// Get UserID from path
		vars := mux.Vars(r)
		userID, ok := vars["userid"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		user, err := s.Klient.GetUser(ctx, userID)
		if err != nil {
			if apierrs.IsNotFound(err) {
				errorResponse := dashv1alpha1.ErrorResponse{
					Message: "user is not found",
				}
				dashv1alpha1.EncodeJSONResponse(errorResponse, pointer.Int(http.StatusNotFound), w)
				return

			} else {
				log.Error(err, "failed to get user", "userid", userID)
				errorResponse := dashv1alpha1.ErrorResponse{
					Message: "failed to get user",
				}
				dashv1alpha1.EncodeJSONResponse(errorResponse, pointer.Int(http.StatusInternalServerError), w)
				return
			}
		}
		ctx = newContextWithUser(ctx, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
