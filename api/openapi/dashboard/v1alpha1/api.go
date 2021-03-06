/*
 * Cosmo Dashboard API
 *
 * Manipulate cosmo dashboard resource API
 *
 * API version: v1alpha1
 * Contact: jlandowner8@gmail.com
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package v1alpha1

import (
	"context"
	"net/http"
)

// AuthApiRouter defines the required methods for binding the api requests to a responses for the AuthApi
// The AuthApiRouter implementation should parse necessary information from the http request,
// pass the data to a AuthApiServicer to perform the required actions, then write the service results to the http response.
type AuthApiRouter interface {
	Login(http.ResponseWriter, *http.Request)
	Logout(http.ResponseWriter, *http.Request)
	Verify(http.ResponseWriter, *http.Request)
}

// TemplateApiRouter defines the required methods for binding the api requests to a responses for the TemplateApi
// The TemplateApiRouter implementation should parse necessary information from the http request,
// pass the data to a TemplateApiServicer to perform the required actions, then write the service results to the http response.
type TemplateApiRouter interface {
	GetUserAddonTemplates(http.ResponseWriter, *http.Request)
	GetWorkspaceTemplates(http.ResponseWriter, *http.Request)
}

// UserApiRouter defines the required methods for binding the api requests to a responses for the UserApi
// The UserApiRouter implementation should parse necessary information from the http request,
// pass the data to a UserApiServicer to perform the required actions, then write the service results to the http response.
type UserApiRouter interface {
	DeleteUser(http.ResponseWriter, *http.Request)
	GetUser(http.ResponseWriter, *http.Request)
	GetUsers(http.ResponseWriter, *http.Request)
	PostUser(http.ResponseWriter, *http.Request)
	PutUserName(http.ResponseWriter, *http.Request)
	PutUserPassword(http.ResponseWriter, *http.Request)
	PutUserRole(http.ResponseWriter, *http.Request)
}

// WorkspaceApiRouter defines the required methods for binding the api requests to a responses for the WorkspaceApi
// The WorkspaceApiRouter implementation should parse necessary information from the http request,
// pass the data to a WorkspaceApiServicer to perform the required actions, then write the service results to the http response.
type WorkspaceApiRouter interface {
	DeleteNetworkRule(http.ResponseWriter, *http.Request)
	DeleteWorkspace(http.ResponseWriter, *http.Request)
	GetWorkspace(http.ResponseWriter, *http.Request)
	GetWorkspaces(http.ResponseWriter, *http.Request)
	PatchWorkspace(http.ResponseWriter, *http.Request)
	PostWorkspace(http.ResponseWriter, *http.Request)
	PutNetworkRule(http.ResponseWriter, *http.Request)
}

// AuthApiServicer defines the api actions for the AuthApi service
// This interface intended to stay up to date with the openapi yaml used to generate it,
// while the service implementation can ignored with the .openapi-generator-ignore file
// and updated with the logic required for the API.
type AuthApiServicer interface {
	Login(context.Context, LoginRequest) (ImplResponse, error)
	Logout(context.Context) (ImplResponse, error)
	Verify(context.Context) (ImplResponse, error)
}

// TemplateApiServicer defines the api actions for the TemplateApi service
// This interface intended to stay up to date with the openapi yaml used to generate it,
// while the service implementation can ignored with the .openapi-generator-ignore file
// and updated with the logic required for the API.
type TemplateApiServicer interface {
	GetUserAddonTemplates(context.Context) (ImplResponse, error)
	GetWorkspaceTemplates(context.Context) (ImplResponse, error)
}

// UserApiServicer defines the api actions for the UserApi service
// This interface intended to stay up to date with the openapi yaml used to generate it,
// while the service implementation can ignored with the .openapi-generator-ignore file
// and updated with the logic required for the API.
type UserApiServicer interface {
	DeleteUser(context.Context, string) (ImplResponse, error)
	GetUser(context.Context, string) (ImplResponse, error)
	GetUsers(context.Context) (ImplResponse, error)
	PostUser(context.Context, CreateUserRequest) (ImplResponse, error)
	PutUserName(context.Context, string, UpdateUserNameRequest) (ImplResponse, error)
	PutUserPassword(context.Context, string, UpdateUserPasswordRequest) (ImplResponse, error)
	PutUserRole(context.Context, string, UpdateUserRoleRequest) (ImplResponse, error)
}

// WorkspaceApiServicer defines the api actions for the WorkspaceApi service
// This interface intended to stay up to date with the openapi yaml used to generate it,
// while the service implementation can ignored with the .openapi-generator-ignore file
// and updated with the logic required for the API.
type WorkspaceApiServicer interface {
	DeleteNetworkRule(context.Context, string, string, string) (ImplResponse, error)
	DeleteWorkspace(context.Context, string, string) (ImplResponse, error)
	GetWorkspace(context.Context, string, string) (ImplResponse, error)
	GetWorkspaces(context.Context, string) (ImplResponse, error)
	PatchWorkspace(context.Context, string, string, PatchWorkspaceRequest) (ImplResponse, error)
	PostWorkspace(context.Context, string, CreateWorkspaceRequest) (ImplResponse, error)
	PutNetworkRule(context.Context, string, string, string, UpsertNetworkRuleRequest) (ImplResponse, error)
}
