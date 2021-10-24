package dashboard

import (
	"context"
	"net/http"

	apierrs "k8s.io/apimachinery/pkg/api/errors"

	dashv1alpha1 "github.com/cosmo-workspace/cosmo/api/openapi/dashboard/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
)

func (s *Server) PutUserRole(ctx context.Context, userId string, req dashv1alpha1.UpdateUserRoleRequest) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "userId", userId, "req", req)

	user := userFromContext(ctx)
	if user == nil {
		log.Info("user not found in context")
		return dashv1alpha1.Response(http.StatusInternalServerError, nil), nil
	}

	userrole := wsv1alpha1.UserRole(req.Role)
	if !userrole.IsValid() {
		log.Info("invalid request", "id", user.ID, "role", userrole)
		return dashv1alpha1.Response(http.StatusBadRequest, nil), nil
	}

	res := &dashv1alpha1.UpdateUserRoleResponse{}

	user.Role = userrole

	updated, err := s.Klient.UpdateUser(ctx, user)
	if err != nil {
		if apierrs.IsNotFound(err) {
			res.Message = err.Error()
			log.Error(err, res.Message, "userid", user.ID)
			return dashv1alpha1.Response(http.StatusNotFound, res), nil
		} else {
			res.Message = "Failed to update user"
			log.Error(err, res.Message, "userid", user.ID)
			return dashv1alpha1.Response(http.StatusInternalServerError, res), nil
		}
	}

	res.User = convertUserToDashv1alpha1User(*updated)
	res.Message = "Successfully updated"
	log.Info(res.Message, "userid", user.ID)
	return dashv1alpha1.Response(http.StatusOK, res), nil
}

func (s *Server) PutUserPassword(ctx context.Context, userId string, req dashv1alpha1.UpdateUserPasswordRequest) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "userId", userId, "req", req)

	user := userFromContext(ctx)
	if user == nil {
		log.Info("user not found in context")
		return dashv1alpha1.Response(http.StatusInternalServerError, nil), nil
	}

	// check current password is valid
	verified, _, err := s.Klient.VerifyPassword(ctx, user.ID, []byte(req.CurrentPassword))
	if err != nil {
		log.Error(err, "failed to get password", "userid", user.ID)
		return dashv1alpha1.Response(http.StatusInternalServerError, nil), nil
	}

	if !verified {
		log.Info("current password invalid", "userid", user.ID)
		return dashv1alpha1.Response(http.StatusForbidden, nil), nil
	}

	res := &dashv1alpha1.UpdateUserPasswordResponse{}

	// Upsert password
	if err := s.Klient.RegisterPassword(ctx, user.ID, []byte(req.NewPassword)); err != nil {
		res.Message = "Failed to update user password"
		log.Error(err, res.Message, "userid", user.ID)
		return dashv1alpha1.Response(http.StatusInternalServerError, res), nil
	}

	res.Message = "Successfully updated"
	log.Info(res.Message, "userid", user.ID)
	return dashv1alpha1.Response(http.StatusOK, res), nil
}
