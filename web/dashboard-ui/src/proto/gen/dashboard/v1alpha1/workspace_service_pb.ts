//
//Cosmo Dashboard API
//Manipulate cosmo dashboard resource API

// @generated by protoc-gen-es v1.2.0 with parameter "target=ts"
// @generated from file dashboard/v1alpha1/workspace_service.proto (package dashboard.v1alpha1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";
import { NetworkRule, Workspace } from "./workspace_pb.js";

/**
 * @generated from message dashboard.v1alpha1.CreateWorkspaceRequest
 */
export class CreateWorkspaceRequest extends Message<CreateWorkspaceRequest> {
  /**
   * @generated from field: string user_name = 1;
   */
  userName = "";

  /**
   * @generated from field: string ws_name = 2;
   */
  wsName = "";

  /**
   * @generated from field: string template = 3;
   */
  template = "";

  /**
   * @generated from field: map<string, string> vars = 4;
   */
  vars: { [key: string]: string } = {};

  constructor(data?: PartialMessage<CreateWorkspaceRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "dashboard.v1alpha1.CreateWorkspaceRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "user_name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "ws_name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "template", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 4, name: "vars", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "scalar", T: 9 /* ScalarType.STRING */} },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CreateWorkspaceRequest {
    return new CreateWorkspaceRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CreateWorkspaceRequest {
    return new CreateWorkspaceRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CreateWorkspaceRequest {
    return new CreateWorkspaceRequest().fromJsonString(jsonString, options);
  }

  static equals(a: CreateWorkspaceRequest | PlainMessage<CreateWorkspaceRequest> | undefined, b: CreateWorkspaceRequest | PlainMessage<CreateWorkspaceRequest> | undefined): boolean {
    return proto3.util.equals(CreateWorkspaceRequest, a, b);
  }
}

/**
 * @generated from message dashboard.v1alpha1.CreateWorkspaceResponse
 */
export class CreateWorkspaceResponse extends Message<CreateWorkspaceResponse> {
  /**
   * @generated from field: string message = 1;
   */
  message = "";

  /**
   * @generated from field: dashboard.v1alpha1.Workspace workspace = 2;
   */
  workspace?: Workspace;

  constructor(data?: PartialMessage<CreateWorkspaceResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "dashboard.v1alpha1.CreateWorkspaceResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "message", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "workspace", kind: "message", T: Workspace },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CreateWorkspaceResponse {
    return new CreateWorkspaceResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CreateWorkspaceResponse {
    return new CreateWorkspaceResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CreateWorkspaceResponse {
    return new CreateWorkspaceResponse().fromJsonString(jsonString, options);
  }

  static equals(a: CreateWorkspaceResponse | PlainMessage<CreateWorkspaceResponse> | undefined, b: CreateWorkspaceResponse | PlainMessage<CreateWorkspaceResponse> | undefined): boolean {
    return proto3.util.equals(CreateWorkspaceResponse, a, b);
  }
}

/**
 * @generated from message dashboard.v1alpha1.DeleteWorkspaceRequest
 */
export class DeleteWorkspaceRequest extends Message<DeleteWorkspaceRequest> {
  /**
   * @generated from field: string user_name = 1;
   */
  userName = "";

  /**
   * @generated from field: string ws_name = 2;
   */
  wsName = "";

  constructor(data?: PartialMessage<DeleteWorkspaceRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "dashboard.v1alpha1.DeleteWorkspaceRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "user_name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "ws_name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DeleteWorkspaceRequest {
    return new DeleteWorkspaceRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DeleteWorkspaceRequest {
    return new DeleteWorkspaceRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DeleteWorkspaceRequest {
    return new DeleteWorkspaceRequest().fromJsonString(jsonString, options);
  }

  static equals(a: DeleteWorkspaceRequest | PlainMessage<DeleteWorkspaceRequest> | undefined, b: DeleteWorkspaceRequest | PlainMessage<DeleteWorkspaceRequest> | undefined): boolean {
    return proto3.util.equals(DeleteWorkspaceRequest, a, b);
  }
}

/**
 * @generated from message dashboard.v1alpha1.DeleteNetworkRuleResponse
 */
export class DeleteNetworkRuleResponse extends Message<DeleteNetworkRuleResponse> {
  /**
   * @generated from field: string message = 1;
   */
  message = "";

  /**
   * @generated from field: dashboard.v1alpha1.NetworkRule network_rule = 2;
   */
  networkRule?: NetworkRule;

  constructor(data?: PartialMessage<DeleteNetworkRuleResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "dashboard.v1alpha1.DeleteNetworkRuleResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "message", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "network_rule", kind: "message", T: NetworkRule },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DeleteNetworkRuleResponse {
    return new DeleteNetworkRuleResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DeleteNetworkRuleResponse {
    return new DeleteNetworkRuleResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DeleteNetworkRuleResponse {
    return new DeleteNetworkRuleResponse().fromJsonString(jsonString, options);
  }

  static equals(a: DeleteNetworkRuleResponse | PlainMessage<DeleteNetworkRuleResponse> | undefined, b: DeleteNetworkRuleResponse | PlainMessage<DeleteNetworkRuleResponse> | undefined): boolean {
    return proto3.util.equals(DeleteNetworkRuleResponse, a, b);
  }
}

/**
 * @generated from message dashboard.v1alpha1.UpdateWorkspaceRequest
 */
export class UpdateWorkspaceRequest extends Message<UpdateWorkspaceRequest> {
  /**
   * @generated from field: string user_name = 1;
   */
  userName = "";

  /**
   * @generated from field: string ws_name = 2;
   */
  wsName = "";

  /**
   * @generated from field: optional int64 replicas = 3;
   */
  replicas?: bigint;

  /**
   * @generated from field: map<string, string> vars = 4;
   */
  vars: { [key: string]: string } = {};

  constructor(data?: PartialMessage<UpdateWorkspaceRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "dashboard.v1alpha1.UpdateWorkspaceRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "user_name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "ws_name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "replicas", kind: "scalar", T: 3 /* ScalarType.INT64 */, opt: true },
    { no: 4, name: "vars", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "scalar", T: 9 /* ScalarType.STRING */} },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): UpdateWorkspaceRequest {
    return new UpdateWorkspaceRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): UpdateWorkspaceRequest {
    return new UpdateWorkspaceRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): UpdateWorkspaceRequest {
    return new UpdateWorkspaceRequest().fromJsonString(jsonString, options);
  }

  static equals(a: UpdateWorkspaceRequest | PlainMessage<UpdateWorkspaceRequest> | undefined, b: UpdateWorkspaceRequest | PlainMessage<UpdateWorkspaceRequest> | undefined): boolean {
    return proto3.util.equals(UpdateWorkspaceRequest, a, b);
  }
}

/**
 * @generated from message dashboard.v1alpha1.UpdateWorkspaceResponse
 */
export class UpdateWorkspaceResponse extends Message<UpdateWorkspaceResponse> {
  /**
   * @generated from field: string message = 1;
   */
  message = "";

  /**
   * @generated from field: dashboard.v1alpha1.Workspace workspace = 2;
   */
  workspace?: Workspace;

  constructor(data?: PartialMessage<UpdateWorkspaceResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "dashboard.v1alpha1.UpdateWorkspaceResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "message", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "workspace", kind: "message", T: Workspace },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): UpdateWorkspaceResponse {
    return new UpdateWorkspaceResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): UpdateWorkspaceResponse {
    return new UpdateWorkspaceResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): UpdateWorkspaceResponse {
    return new UpdateWorkspaceResponse().fromJsonString(jsonString, options);
  }

  static equals(a: UpdateWorkspaceResponse | PlainMessage<UpdateWorkspaceResponse> | undefined, b: UpdateWorkspaceResponse | PlainMessage<UpdateWorkspaceResponse> | undefined): boolean {
    return proto3.util.equals(UpdateWorkspaceResponse, a, b);
  }
}

/**
 * @generated from message dashboard.v1alpha1.GetWorkspaceRequest
 */
export class GetWorkspaceRequest extends Message<GetWorkspaceRequest> {
  /**
   * @generated from field: string user_name = 1;
   */
  userName = "";

  /**
   * @generated from field: string ws_name = 2;
   */
  wsName = "";

  /**
   * @generated from field: optional bool with_raw = 3;
   */
  withRaw?: boolean;

  constructor(data?: PartialMessage<GetWorkspaceRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "dashboard.v1alpha1.GetWorkspaceRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "user_name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "ws_name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "with_raw", kind: "scalar", T: 8 /* ScalarType.BOOL */, opt: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetWorkspaceRequest {
    return new GetWorkspaceRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetWorkspaceRequest {
    return new GetWorkspaceRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetWorkspaceRequest {
    return new GetWorkspaceRequest().fromJsonString(jsonString, options);
  }

  static equals(a: GetWorkspaceRequest | PlainMessage<GetWorkspaceRequest> | undefined, b: GetWorkspaceRequest | PlainMessage<GetWorkspaceRequest> | undefined): boolean {
    return proto3.util.equals(GetWorkspaceRequest, a, b);
  }
}

/**
 * @generated from message dashboard.v1alpha1.GetWorkspaceResponse
 */
export class GetWorkspaceResponse extends Message<GetWorkspaceResponse> {
  /**
   * @generated from field: dashboard.v1alpha1.Workspace workspace = 1;
   */
  workspace?: Workspace;

  constructor(data?: PartialMessage<GetWorkspaceResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "dashboard.v1alpha1.GetWorkspaceResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "workspace", kind: "message", T: Workspace },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetWorkspaceResponse {
    return new GetWorkspaceResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetWorkspaceResponse {
    return new GetWorkspaceResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetWorkspaceResponse {
    return new GetWorkspaceResponse().fromJsonString(jsonString, options);
  }

  static equals(a: GetWorkspaceResponse | PlainMessage<GetWorkspaceResponse> | undefined, b: GetWorkspaceResponse | PlainMessage<GetWorkspaceResponse> | undefined): boolean {
    return proto3.util.equals(GetWorkspaceResponse, a, b);
  }
}

/**
 * @generated from message dashboard.v1alpha1.GetWorkspacesRequest
 */
export class GetWorkspacesRequest extends Message<GetWorkspacesRequest> {
  /**
   * @generated from field: string user_name = 1;
   */
  userName = "";

  /**
   * @generated from field: optional bool with_raw = 2;
   */
  withRaw?: boolean;

  /**
   * @generated from field: optional bool includeShared = 3;
   */
  includeShared?: boolean;

  constructor(data?: PartialMessage<GetWorkspacesRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "dashboard.v1alpha1.GetWorkspacesRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "user_name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "with_raw", kind: "scalar", T: 8 /* ScalarType.BOOL */, opt: true },
    { no: 3, name: "includeShared", kind: "scalar", T: 8 /* ScalarType.BOOL */, opt: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetWorkspacesRequest {
    return new GetWorkspacesRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetWorkspacesRequest {
    return new GetWorkspacesRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetWorkspacesRequest {
    return new GetWorkspacesRequest().fromJsonString(jsonString, options);
  }

  static equals(a: GetWorkspacesRequest | PlainMessage<GetWorkspacesRequest> | undefined, b: GetWorkspacesRequest | PlainMessage<GetWorkspacesRequest> | undefined): boolean {
    return proto3.util.equals(GetWorkspacesRequest, a, b);
  }
}

/**
 * @generated from message dashboard.v1alpha1.GetWorkspacesResponse
 */
export class GetWorkspacesResponse extends Message<GetWorkspacesResponse> {
  /**
   * @generated from field: string message = 1;
   */
  message = "";

  /**
   * @generated from field: repeated dashboard.v1alpha1.Workspace items = 2;
   */
  items: Workspace[] = [];

  constructor(data?: PartialMessage<GetWorkspacesResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "dashboard.v1alpha1.GetWorkspacesResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "message", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "items", kind: "message", T: Workspace, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetWorkspacesResponse {
    return new GetWorkspacesResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetWorkspacesResponse {
    return new GetWorkspacesResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetWorkspacesResponse {
    return new GetWorkspacesResponse().fromJsonString(jsonString, options);
  }

  static equals(a: GetWorkspacesResponse | PlainMessage<GetWorkspacesResponse> | undefined, b: GetWorkspacesResponse | PlainMessage<GetWorkspacesResponse> | undefined): boolean {
    return proto3.util.equals(GetWorkspacesResponse, a, b);
  }
}

/**
 * @generated from message dashboard.v1alpha1.UpsertNetworkRuleRequest
 */
export class UpsertNetworkRuleRequest extends Message<UpsertNetworkRuleRequest> {
  /**
   * @generated from field: string user_name = 1;
   */
  userName = "";

  /**
   * @generated from field: string ws_name = 2;
   */
  wsName = "";

  /**
   * @generated from field: dashboard.v1alpha1.NetworkRule network_rule = 3;
   */
  networkRule?: NetworkRule;

  /**
   * network rule index to update. insert if index is out of length
   *
   * @generated from field: int32 index = 4;
   */
  index = 0;

  constructor(data?: PartialMessage<UpsertNetworkRuleRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "dashboard.v1alpha1.UpsertNetworkRuleRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "user_name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "ws_name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "network_rule", kind: "message", T: NetworkRule },
    { no: 4, name: "index", kind: "scalar", T: 5 /* ScalarType.INT32 */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): UpsertNetworkRuleRequest {
    return new UpsertNetworkRuleRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): UpsertNetworkRuleRequest {
    return new UpsertNetworkRuleRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): UpsertNetworkRuleRequest {
    return new UpsertNetworkRuleRequest().fromJsonString(jsonString, options);
  }

  static equals(a: UpsertNetworkRuleRequest | PlainMessage<UpsertNetworkRuleRequest> | undefined, b: UpsertNetworkRuleRequest | PlainMessage<UpsertNetworkRuleRequest> | undefined): boolean {
    return proto3.util.equals(UpsertNetworkRuleRequest, a, b);
  }
}

/**
 * @generated from message dashboard.v1alpha1.UpsertNetworkRuleResponse
 */
export class UpsertNetworkRuleResponse extends Message<UpsertNetworkRuleResponse> {
  /**
   * @generated from field: string message = 1;
   */
  message = "";

  /**
   * @generated from field: dashboard.v1alpha1.NetworkRule network_rule = 2;
   */
  networkRule?: NetworkRule;

  constructor(data?: PartialMessage<UpsertNetworkRuleResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "dashboard.v1alpha1.UpsertNetworkRuleResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "message", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "network_rule", kind: "message", T: NetworkRule },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): UpsertNetworkRuleResponse {
    return new UpsertNetworkRuleResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): UpsertNetworkRuleResponse {
    return new UpsertNetworkRuleResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): UpsertNetworkRuleResponse {
    return new UpsertNetworkRuleResponse().fromJsonString(jsonString, options);
  }

  static equals(a: UpsertNetworkRuleResponse | PlainMessage<UpsertNetworkRuleResponse> | undefined, b: UpsertNetworkRuleResponse | PlainMessage<UpsertNetworkRuleResponse> | undefined): boolean {
    return proto3.util.equals(UpsertNetworkRuleResponse, a, b);
  }
}

/**
 * @generated from message dashboard.v1alpha1.DeleteNetworkRuleRequest
 */
export class DeleteNetworkRuleRequest extends Message<DeleteNetworkRuleRequest> {
  /**
   * @generated from field: string user_name = 1;
   */
  userName = "";

  /**
   * @generated from field: string ws_name = 2;
   */
  wsName = "";

  /**
   * @generated from field: int32 index = 3;
   */
  index = 0;

  constructor(data?: PartialMessage<DeleteNetworkRuleRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "dashboard.v1alpha1.DeleteNetworkRuleRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "user_name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "ws_name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "index", kind: "scalar", T: 5 /* ScalarType.INT32 */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DeleteNetworkRuleRequest {
    return new DeleteNetworkRuleRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DeleteNetworkRuleRequest {
    return new DeleteNetworkRuleRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DeleteNetworkRuleRequest {
    return new DeleteNetworkRuleRequest().fromJsonString(jsonString, options);
  }

  static equals(a: DeleteNetworkRuleRequest | PlainMessage<DeleteNetworkRuleRequest> | undefined, b: DeleteNetworkRuleRequest | PlainMessage<DeleteNetworkRuleRequest> | undefined): boolean {
    return proto3.util.equals(DeleteNetworkRuleRequest, a, b);
  }
}

/**
 * @generated from message dashboard.v1alpha1.DeleteWorkspaceResponse
 */
export class DeleteWorkspaceResponse extends Message<DeleteWorkspaceResponse> {
  /**
   * @generated from field: string message = 1;
   */
  message = "";

  /**
   * @generated from field: dashboard.v1alpha1.Workspace workspace = 2;
   */
  workspace?: Workspace;

  constructor(data?: PartialMessage<DeleteWorkspaceResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "dashboard.v1alpha1.DeleteWorkspaceResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "message", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "workspace", kind: "message", T: Workspace },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DeleteWorkspaceResponse {
    return new DeleteWorkspaceResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DeleteWorkspaceResponse {
    return new DeleteWorkspaceResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DeleteWorkspaceResponse {
    return new DeleteWorkspaceResponse().fromJsonString(jsonString, options);
  }

  static equals(a: DeleteWorkspaceResponse | PlainMessage<DeleteWorkspaceResponse> | undefined, b: DeleteWorkspaceResponse | PlainMessage<DeleteWorkspaceResponse> | undefined): boolean {
    return proto3.util.equals(DeleteWorkspaceResponse, a, b);
  }
}

