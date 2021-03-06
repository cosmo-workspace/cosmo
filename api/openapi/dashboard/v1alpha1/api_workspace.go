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
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// WorkspaceApiController binds http requests to an api service and writes the service results to the http response
type WorkspaceApiController struct {
	service      WorkspaceApiServicer
	errorHandler ErrorHandler
}

// WorkspaceApiOption for how the controller is set up.
type WorkspaceApiOption func(*WorkspaceApiController)

// WithWorkspaceApiErrorHandler inject ErrorHandler into controller
func WithWorkspaceApiErrorHandler(h ErrorHandler) WorkspaceApiOption {
	return func(c *WorkspaceApiController) {
		c.errorHandler = h
	}
}

// NewWorkspaceApiController creates a default api controller
func NewWorkspaceApiController(s WorkspaceApiServicer, opts ...WorkspaceApiOption) Router {
	controller := &WorkspaceApiController{
		service:      s,
		errorHandler: DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(controller)
	}

	return controller
}

// Routes returns all of the api route for the WorkspaceApiController
func (c *WorkspaceApiController) Routes() Routes {
	return Routes{
		{
			"DeleteNetworkRule",
			strings.ToUpper("Delete"),
			"/api/v1alpha1/user/{userid}/workspace/{wsName}/network/{networkRuleName}",
			c.DeleteNetworkRule,
		},
		{
			"DeleteWorkspace",
			strings.ToUpper("Delete"),
			"/api/v1alpha1/user/{userid}/workspace/{wsName}",
			c.DeleteWorkspace,
		},
		{
			"GetWorkspace",
			strings.ToUpper("Get"),
			"/api/v1alpha1/user/{userid}/workspace/{wsName}",
			c.GetWorkspace,
		},
		{
			"GetWorkspaces",
			strings.ToUpper("Get"),
			"/api/v1alpha1/user/{userid}/workspace",
			c.GetWorkspaces,
		},
		{
			"PatchWorkspace",
			strings.ToUpper("Patch"),
			"/api/v1alpha1/user/{userid}/workspace/{wsName}",
			c.PatchWorkspace,
		},
		{
			"PostWorkspace",
			strings.ToUpper("Post"),
			"/api/v1alpha1/user/{userid}/workspace",
			c.PostWorkspace,
		},
		{
			"PutNetworkRule",
			strings.ToUpper("Put"),
			"/api/v1alpha1/user/{userid}/workspace/{wsName}/network/{networkRuleName}",
			c.PutNetworkRule,
		},
	}
}

// DeleteNetworkRule - Remove workspace network rule
func (c *WorkspaceApiController) DeleteNetworkRule(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	useridParam := params["userid"]

	wsNameParam := params["wsName"]

	networkRuleNameParam := params["networkRuleName"]

	result, err := c.service.DeleteNetworkRule(r.Context(), useridParam, wsNameParam, networkRuleNameParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// DeleteWorkspace - Delete workspace.
func (c *WorkspaceApiController) DeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	useridParam := params["userid"]

	wsNameParam := params["wsName"]

	result, err := c.service.DeleteWorkspace(r.Context(), useridParam, wsNameParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// GetWorkspace - Get workspace by name.
func (c *WorkspaceApiController) GetWorkspace(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	useridParam := params["userid"]

	wsNameParam := params["wsName"]

	result, err := c.service.GetWorkspace(r.Context(), useridParam, wsNameParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// GetWorkspaces - Get all workspace of user.
func (c *WorkspaceApiController) GetWorkspaces(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	useridParam := params["userid"]

	result, err := c.service.GetWorkspaces(r.Context(), useridParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// PatchWorkspace - Update workspace.
func (c *WorkspaceApiController) PatchWorkspace(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	useridParam := params["userid"]

	wsNameParam := params["wsName"]

	patchWorkspaceRequestParam := PatchWorkspaceRequest{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&patchWorkspaceRequestParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertPatchWorkspaceRequestRequired(patchWorkspaceRequestParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.PatchWorkspace(r.Context(), useridParam, wsNameParam, patchWorkspaceRequestParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// PostWorkspace - Create a new Workspace
func (c *WorkspaceApiController) PostWorkspace(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	useridParam := params["userid"]

	createWorkspaceRequestParam := CreateWorkspaceRequest{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&createWorkspaceRequestParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertCreateWorkspaceRequestRequired(createWorkspaceRequestParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.PostWorkspace(r.Context(), useridParam, createWorkspaceRequestParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// PutNetworkRule - Upsert workspace network rule
func (c *WorkspaceApiController) PutNetworkRule(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	useridParam := params["userid"]

	wsNameParam := params["wsName"]

	networkRuleNameParam := params["networkRuleName"]

	upsertNetworkRuleRequestParam := UpsertNetworkRuleRequest{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&upsertNetworkRuleRequestParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertUpsertNetworkRuleRequestRequired(upsertNetworkRuleRequestParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.PutNetworkRule(r.Context(), useridParam, wsNameParam, networkRuleNameParam, upsertNetworkRuleRequestParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}
