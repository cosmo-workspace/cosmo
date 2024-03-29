//
//Cosmo Dashboard API
//Manipulate cosmo dashboard resource API

// @generated by protoc-gen-connect-web v0.11.0 with parameter "target=ts"
// @generated from file dashboard/v1alpha1/workspace_service.proto (package dashboard.v1alpha1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { CreateWorkspaceRequest, CreateWorkspaceResponse, DeleteNetworkRuleRequest, DeleteNetworkRuleResponse, DeleteWorkspaceRequest, DeleteWorkspaceResponse, GetWorkspaceRequest, GetWorkspaceResponse, GetWorkspacesRequest, GetWorkspacesResponse, UpdateWorkspaceRequest, UpdateWorkspaceResponse, UpsertNetworkRuleRequest, UpsertNetworkRuleResponse } from "./workspace_service_pb.js";
import { MethodKind } from "@bufbuild/protobuf";

/**
 * @generated from service dashboard.v1alpha1.WorkspaceService
 */
export const WorkspaceService = {
  typeName: "dashboard.v1alpha1.WorkspaceService",
  methods: {
    /**
     * Create a new Workspace
     *
     * @generated from rpc dashboard.v1alpha1.WorkspaceService.CreateWorkspace
     */
    createWorkspace: {
      name: "CreateWorkspace",
      I: CreateWorkspaceRequest,
      O: CreateWorkspaceResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Delete workspace
     *
     * @generated from rpc dashboard.v1alpha1.WorkspaceService.DeleteWorkspace
     */
    deleteWorkspace: {
      name: "DeleteWorkspace",
      I: DeleteWorkspaceRequest,
      O: DeleteWorkspaceResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Update workspace
     *
     * @generated from rpc dashboard.v1alpha1.WorkspaceService.UpdateWorkspace
     */
    updateWorkspace: {
      name: "UpdateWorkspace",
      I: UpdateWorkspaceRequest,
      O: UpdateWorkspaceResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Returns a single Workspace model
     *
     * @generated from rpc dashboard.v1alpha1.WorkspaceService.GetWorkspace
     */
    getWorkspace: {
      name: "GetWorkspace",
      I: GetWorkspaceRequest,
      O: GetWorkspaceResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Returns an array of Workspace model
     *
     * @generated from rpc dashboard.v1alpha1.WorkspaceService.GetWorkspaces
     */
    getWorkspaces: {
      name: "GetWorkspaces",
      I: GetWorkspacesRequest,
      O: GetWorkspacesResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Upsert workspace network rule
     *
     * @generated from rpc dashboard.v1alpha1.WorkspaceService.UpsertNetworkRule
     */
    upsertNetworkRule: {
      name: "UpsertNetworkRule",
      I: UpsertNetworkRuleRequest,
      O: UpsertNetworkRuleResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Remove workspace network rule
     *
     * @generated from rpc dashboard.v1alpha1.WorkspaceService.DeleteNetworkRule
     */
    deleteNetworkRule: {
      name: "DeleteNetworkRule",
      I: DeleteNetworkRuleRequest,
      O: DeleteNetworkRuleResponse,
      kind: MethodKind.Unary,
    },
  }
} as const;

