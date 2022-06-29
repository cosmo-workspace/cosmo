package dashboard

import (
	"context"
	"errors"
	"net/http"

	dashv1alpha1 "github.com/cosmo-workspace/cosmo/api/openapi/dashboard/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
)

func (s *Server) PutUserName(ctx context.Context, userId string, req dashv1alpha1.UpdateUserNameRequest) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "userId", userId, "req", req)

	user, err := s.Klient.UpdateUser(ctx, userId, kosmo.UpdateUserOpts{DisplayName: &req.DisplayName})
	if err != nil {
		return ErrorResponse(log, err)
	}

	res := &dashv1alpha1.UpdateUserNameResponse{
		Message: "Successfully updated",
		User:    convertUserToDashv1alpha1User(*user),
	}
	log.Info(res.Message, "userid", userId)
	return NormalResponse(http.StatusOK, res)
}

func (s *Server) PutUserRole(ctx context.Context, userId string, req dashv1alpha1.UpdateUserRoleRequest) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "userId", userId, "req", req)

	user, err := s.Klient.UpdateUser(ctx, userId, kosmo.UpdateUserOpts{UserRole: &req.Role})
	if err != nil {
		return ErrorResponse(log, err)
	}

	res := &dashv1alpha1.UpdateUserNameResponse{
		Message: "Successfully updated",
		User:    convertUserToDashv1alpha1User(*user),
	}
	log.Info(res.Message, "userid", userId)
	return NormalResponse(http.StatusOK, res)
}

func (s *Server) PutUserPassword(ctx context.Context, userId string, req dashv1alpha1.UpdateUserPasswordRequest) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "userId", userId, "req", req)

	// check current password is valid
	verified, _, err := s.Klient.VerifyPassword(ctx, userId, []byte(req.CurrentPassword))
	if err != nil {
		//todo
		if errors.Is(err, kosmo.ErrNotFound) {
			return ErrorResponse(log, err)
		} else {
			return ErrorResponse(log, kosmo.NewInternalServerError("", err))
		}
	}

	if !verified {
		err := kosmo.NewForbiddenError("incorrect user or password", err)
		return ErrorResponse(log, err)
	}

	// Upsert password
	if err := s.Klient.RegisterPassword(ctx, userId, []byte(req.NewPassword)); err != nil {
		return ErrorResponse(log, err)
	}

	res := &dashv1alpha1.UpdateUserPasswordResponse{
		Message: "Successfully updated",
	}
	log.Info(res.Message, "userid", userId)
	return NormalResponse(http.StatusOK, res)
}
