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

type ctxKeyWorkspace struct{}

func newContextWithWorkspace(ctx context.Context, ws *wsv1alpha1.Workspace) context.Context {
	return context.WithValue(ctx, ctxKeyWorkspace{}, ws)
}

func workspaceFromContext(ctx context.Context) *wsv1alpha1.Workspace {
	ws, ok := ctx.Value(ctxKeyWorkspace{}).(*wsv1alpha1.Workspace)
	if ok && ws != nil {
		return ws.DeepCopy()
	}
	return nil
}

func (s *Server) preFetchWorkspaceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := clog.FromContext(ctx).WithName("preFetchWorkspace")

		vars := mux.Vars(r)

		// Get UserID from path
		userID, ok := vars["userid"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Get Workspace name from path
		wsName, ok := vars["wsName"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ws, err := s.Klient.GetWorkspaceByUserID(ctx, wsName, userID)
		if err != nil {
			if apierrs.IsNotFound(err) {
				errorResponse := dashv1alpha1.ErrorResponse{
					Message: "workspace is not found",
				}
				dashv1alpha1.EncodeJSONResponse(errorResponse, pointer.Int(http.StatusNotFound), w)
				return

			} else {
				log.Error(err, "failed to get workspace", "userid", userID, "workspace", wsName)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		log.DebugAll().Info("save ws in ctx")
		ctx = newContextWithWorkspace(ctx, ws)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
