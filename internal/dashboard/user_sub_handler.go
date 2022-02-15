package dashboard

import (
	"context"
	"net/http"

	apierrs "k8s.io/apimachinery/pkg/api/errors"

	dashv1alpha1 "github.com/cosmo-workspace/cosmo/api/openapi/dashboard/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
)

func (s *Server) PutUserName(ctx context.Context, userId string, req dashv1alpha1.UpdateUserNameRequest) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "userId", userId, "req", req)

	user := userFromContext(ctx)
	if user == nil {
		log.Info("user not found in context")
		return ErrorResponse(http.StatusInternalServerError, "")
	}

	user.Spec.DisplayName = req.DisplayName

	err := s.Klient.Update(ctx, user)
	if err != nil {
		if apierrs.IsNotFound(err) {
			log.Error(err, err.Error(), "userid", user.Name)
			return ErrorResponse(http.StatusNotFound, err.Error())
		} else {
			message := "Failed to update user"
			log.Error(err, message, "userid", user.Name)
			return ErrorResponse(http.StatusInternalServerError, message)
		}
	}

	res := &dashv1alpha1.UpdateUserNameResponse{}
	res.User = convertUserToDashv1alpha1User(*user)
	res.Message = "Successfully updated"
	log.Info(res.Message, "userid", user.Name)
	return NormalResponse(http.StatusOK, res)
}

func (s *Server) PutUserRole(ctx context.Context, userId string, req dashv1alpha1.UpdateUserRoleRequest) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "userId", userId, "req", req)

	user := userFromContext(ctx)
	if user == nil {
		log.Info("user not found in context")
		return ErrorResponse(http.StatusInternalServerError, "")
	}

	userrole := wsv1alpha1.UserRole(req.Role)
	if !userrole.IsValid() {
		log.Info("invalid request", "id", user.Name, "role", userrole)
		return ErrorResponse(http.StatusBadRequest, "")
	}
	user.Spec.Role = userrole

	err := s.Klient.Update(ctx, user)
	if err != nil {
		if apierrs.IsNotFound(err) {
			log.Error(err, err.Error(), "userid", user.Name)
			return ErrorResponse(http.StatusNotFound, err.Error())
		} else {
			message := "Failed to update user"
			log.Error(err, message, "userid", user.Name)
			return ErrorResponse(http.StatusInternalServerError, message)
		}
	}

	res := &dashv1alpha1.UpdateUserNameResponse{}
	res.User = convertUserToDashv1alpha1User(*user)
	res.Message = "Successfully updated"
	log.Info(res.Message, "userid", user.Name)
	return NormalResponse(http.StatusOK, res)
}

func (s *Server) PutUserPassword(ctx context.Context, userId string, req dashv1alpha1.UpdateUserPasswordRequest) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "userId", userId, "req", req)

	user := userFromContext(ctx)
	if user == nil {
		log.Info("user not found in context")
		return ErrorResponse(http.StatusInternalServerError, "")
	}

	// check current password is valid
	verified, _, err := s.Klient.VerifyPassword(ctx, user.Name, []byte(req.CurrentPassword))
	if err != nil {
		log.Error(err, "failed to get password", "userid", user.Name)
		return ErrorResponse(http.StatusInternalServerError, "")
	}

	if !verified {
		log.Info("current password invalid", "userid", user.Name)
		return ErrorResponse(http.StatusForbidden, "")
	}

	// Upsert password
	if err := s.Klient.RegisterPassword(ctx, user.Name, []byte(req.NewPassword)); err != nil {
		message := "Failed to update user password"
		log.Error(err, message, "userid", user.Name)
		return ErrorResponse(http.StatusInternalServerError, message)
	}

	res := &dashv1alpha1.UpdateUserPasswordResponse{}
	res.Message = "Successfully updated"
	log.Info(res.Message, "userid", user.Name)
	return NormalResponse(http.StatusOK, res)
}
