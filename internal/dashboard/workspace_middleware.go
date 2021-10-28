package dashboard

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	apierrs "k8s.io/apimachinery/pkg/api/errors"

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

		user := userFromContext(ctx)
		if user == nil {
			log.Info("user not found in context")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Get Workspace name from path
		vars := mux.Vars(r)
		wsName, ok := vars["wsName"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ws, err := s.Klient.GetWorkspaceByUserID(ctx, wsName, user.Name)
		if err != nil {
			if apierrs.IsNotFound(err) {
				w.WriteHeader(http.StatusNotFound)
				return

			} else {
				log.Error(err, "failed to get workspace", "userid", user.Name, "workspace", wsName)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		log.DebugAll().Info("save ws in ctx")
		ctx = newContextWithWorkspace(ctx, ws)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
