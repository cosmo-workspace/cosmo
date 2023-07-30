# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [dashboard/v1alpha1/auth_service.proto](#dashboard_v1alpha1_auth_service-proto)
    - [LoginRequest](#dashboard-v1alpha1-LoginRequest)
    - [LoginResponse](#dashboard-v1alpha1-LoginResponse)
    - [VerifyResponse](#dashboard-v1alpha1-VerifyResponse)
  
    - [AuthService](#dashboard-v1alpha1-AuthService)
  
- [dashboard/v1alpha1/template.proto](#dashboard_v1alpha1_template-proto)
    - [Template](#dashboard-v1alpha1-Template)
    - [TemplateRequiredVars](#dashboard-v1alpha1-TemplateRequiredVars)
  
- [dashboard/v1alpha1/template_service.proto](#dashboard_v1alpha1_template_service-proto)
    - [GetUserAddonTemplatesResponse](#dashboard-v1alpha1-GetUserAddonTemplatesResponse)
    - [GetWorkspaceTemplatesResponse](#dashboard-v1alpha1-GetWorkspaceTemplatesResponse)
  
    - [TemplateService](#dashboard-v1alpha1-TemplateService)
  
- [dashboard/v1alpha1/user.proto](#dashboard_v1alpha1_user-proto)
    - [User](#dashboard-v1alpha1-User)
    - [UserAddon](#dashboard-v1alpha1-UserAddon)
    - [UserAddon.VarsEntry](#dashboard-v1alpha1-UserAddon-VarsEntry)
  
- [dashboard/v1alpha1/user_service.proto](#dashboard_v1alpha1_user_service-proto)
    - [CreateUserRequest](#dashboard-v1alpha1-CreateUserRequest)
    - [CreateUserResponse](#dashboard-v1alpha1-CreateUserResponse)
    - [DeleteUserRequest](#dashboard-v1alpha1-DeleteUserRequest)
    - [DeleteUserResponse](#dashboard-v1alpha1-DeleteUserResponse)
    - [GetUserRequest](#dashboard-v1alpha1-GetUserRequest)
    - [GetUserResponse](#dashboard-v1alpha1-GetUserResponse)
    - [GetUsersResponse](#dashboard-v1alpha1-GetUsersResponse)
    - [UpdateUserAddonsRequest](#dashboard-v1alpha1-UpdateUserAddonsRequest)
    - [UpdateUserAddonsResponse](#dashboard-v1alpha1-UpdateUserAddonsResponse)
    - [UpdateUserDisplayNameRequest](#dashboard-v1alpha1-UpdateUserDisplayNameRequest)
    - [UpdateUserDisplayNameResponse](#dashboard-v1alpha1-UpdateUserDisplayNameResponse)
    - [UpdateUserPasswordRequest](#dashboard-v1alpha1-UpdateUserPasswordRequest)
    - [UpdateUserPasswordResponse](#dashboard-v1alpha1-UpdateUserPasswordResponse)
    - [UpdateUserRoleRequest](#dashboard-v1alpha1-UpdateUserRoleRequest)
    - [UpdateUserRoleResponse](#dashboard-v1alpha1-UpdateUserRoleResponse)
  
    - [UserService](#dashboard-v1alpha1-UserService)
  
- [dashboard/v1alpha1/workspace.proto](#dashboard_v1alpha1_workspace-proto)
    - [NetworkRule](#dashboard-v1alpha1-NetworkRule)
    - [Workspace](#dashboard-v1alpha1-Workspace)
    - [WorkspaceSpec](#dashboard-v1alpha1-WorkspaceSpec)
    - [WorkspaceSpec.VarsEntry](#dashboard-v1alpha1-WorkspaceSpec-VarsEntry)
    - [WorkspaceStatus](#dashboard-v1alpha1-WorkspaceStatus)
  
- [dashboard/v1alpha1/workspace_service.proto](#dashboard_v1alpha1_workspace_service-proto)
    - [CreateWorkspaceRequest](#dashboard-v1alpha1-CreateWorkspaceRequest)
    - [CreateWorkspaceRequest.VarsEntry](#dashboard-v1alpha1-CreateWorkspaceRequest-VarsEntry)
    - [CreateWorkspaceResponse](#dashboard-v1alpha1-CreateWorkspaceResponse)
    - [DeleteNetworkRuleRequest](#dashboard-v1alpha1-DeleteNetworkRuleRequest)
    - [DeleteNetworkRuleResponse](#dashboard-v1alpha1-DeleteNetworkRuleResponse)
    - [DeleteWorkspaceRequest](#dashboard-v1alpha1-DeleteWorkspaceRequest)
    - [DeleteWorkspaceResponse](#dashboard-v1alpha1-DeleteWorkspaceResponse)
    - [GetWorkspaceRequest](#dashboard-v1alpha1-GetWorkspaceRequest)
    - [GetWorkspaceResponse](#dashboard-v1alpha1-GetWorkspaceResponse)
    - [GetWorkspacesRequest](#dashboard-v1alpha1-GetWorkspacesRequest)
    - [GetWorkspacesResponse](#dashboard-v1alpha1-GetWorkspacesResponse)
    - [UpdateWorkspaceRequest](#dashboard-v1alpha1-UpdateWorkspaceRequest)
    - [UpdateWorkspaceResponse](#dashboard-v1alpha1-UpdateWorkspaceResponse)
    - [UpsertNetworkRuleRequest](#dashboard-v1alpha1-UpsertNetworkRuleRequest)
    - [UpsertNetworkRuleResponse](#dashboard-v1alpha1-UpsertNetworkRuleResponse)
  
    - [WorkspaceService](#dashboard-v1alpha1-WorkspaceService)
  
- [Scalar Value Types](#scalar-value-types)



<a name="dashboard_v1alpha1_auth_service-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## dashboard/v1alpha1/auth_service.proto



<a name="dashboard-v1alpha1-LoginRequest"></a>

### LoginRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_name | [string](#string) |  |  |
| password | [string](#string) |  |  |






<a name="dashboard-v1alpha1-LoginResponse"></a>

### LoginResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_name | [string](#string) |  |  |
| expire_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| require_password_update | [bool](#bool) |  |  |






<a name="dashboard-v1alpha1-VerifyResponse"></a>

### VerifyResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_name | [string](#string) |  |  |
| expire_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| require_password_update | [bool](#bool) |  |  |





 

 

 


<a name="dashboard-v1alpha1-AuthService"></a>

### AuthService


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Login | [LoginRequest](#dashboard-v1alpha1-LoginRequest) | [LoginResponse](#dashboard-v1alpha1-LoginResponse) | ID and password to login |
| Logout | [.google.protobuf.Empty](#google-protobuf-Empty) | [.google.protobuf.Empty](#google-protobuf-Empty) | Delete session to logout |
| Verify | [.google.protobuf.Empty](#google-protobuf-Empty) | [VerifyResponse](#dashboard-v1alpha1-VerifyResponse) | Verify authorization |

 



<a name="dashboard_v1alpha1_template-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## dashboard/v1alpha1/template.proto



<a name="dashboard-v1alpha1-Template"></a>

### Template



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| description | [string](#string) |  |  |
| required_vars | [TemplateRequiredVars](#dashboard-v1alpha1-TemplateRequiredVars) | repeated |  |
| is_default_user_addon | [bool](#bool) | optional |  |
| is_cluster_scope | [bool](#bool) |  |  |






<a name="dashboard-v1alpha1-TemplateRequiredVars"></a>

### TemplateRequiredVars



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| var_name | [string](#string) |  |  |
| default_value | [string](#string) |  |  |





 

 

 

 



<a name="dashboard_v1alpha1_template_service-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## dashboard/v1alpha1/template_service.proto



<a name="dashboard-v1alpha1-GetUserAddonTemplatesResponse"></a>

### GetUserAddonTemplatesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| items | [Template](#dashboard-v1alpha1-Template) | repeated |  |






<a name="dashboard-v1alpha1-GetWorkspaceTemplatesResponse"></a>

### GetWorkspaceTemplatesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| items | [Template](#dashboard-v1alpha1-Template) | repeated |  |





 

 

 


<a name="dashboard-v1alpha1-TemplateService"></a>

### TemplateService


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetUserAddonTemplates | [.google.protobuf.Empty](#google-protobuf-Empty) | [GetUserAddonTemplatesResponse](#dashboard-v1alpha1-GetUserAddonTemplatesResponse) | List templates typed useraddon |
| GetWorkspaceTemplates | [.google.protobuf.Empty](#google-protobuf-Empty) | [GetWorkspaceTemplatesResponse](#dashboard-v1alpha1-GetWorkspaceTemplatesResponse) | List templates typed workspace |

 



<a name="dashboard_v1alpha1_user-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## dashboard/v1alpha1/user.proto



<a name="dashboard-v1alpha1-User"></a>

### User



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| display_name | [string](#string) |  |  |
| roles | [string](#string) | repeated |  |
| auth_type | [string](#string) |  |  |
| addons | [UserAddon](#dashboard-v1alpha1-UserAddon) | repeated |  |
| default_password | [string](#string) |  |  |
| status | [string](#string) |  |  |






<a name="dashboard-v1alpha1-UserAddon"></a>

### UserAddon



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| template | [string](#string) |  |  |
| cluster_scoped | [bool](#bool) |  |  |
| vars | [UserAddon.VarsEntry](#dashboard-v1alpha1-UserAddon-VarsEntry) | repeated |  |






<a name="dashboard-v1alpha1-UserAddon-VarsEntry"></a>

### UserAddon.VarsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |





 

 

 

 



<a name="dashboard_v1alpha1_user_service-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## dashboard/v1alpha1/user_service.proto



<a name="dashboard-v1alpha1-CreateUserRequest"></a>

### CreateUserRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_name | [string](#string) |  |  |
| display_name | [string](#string) |  |  |
| roles | [string](#string) | repeated |  |
| auth_type | [string](#string) |  |  |
| addons | [UserAddon](#dashboard-v1alpha1-UserAddon) | repeated |  |






<a name="dashboard-v1alpha1-CreateUserResponse"></a>

### CreateUserResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| user | [User](#dashboard-v1alpha1-User) |  |  |






<a name="dashboard-v1alpha1-DeleteUserRequest"></a>

### DeleteUserRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_name | [string](#string) |  |  |






<a name="dashboard-v1alpha1-DeleteUserResponse"></a>

### DeleteUserResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| user | [User](#dashboard-v1alpha1-User) |  |  |






<a name="dashboard-v1alpha1-GetUserRequest"></a>

### GetUserRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_name | [string](#string) |  |  |






<a name="dashboard-v1alpha1-GetUserResponse"></a>

### GetUserResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user | [User](#dashboard-v1alpha1-User) |  |  |






<a name="dashboard-v1alpha1-GetUsersResponse"></a>

### GetUsersResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| items | [User](#dashboard-v1alpha1-User) | repeated |  |






<a name="dashboard-v1alpha1-UpdateUserAddonsRequest"></a>

### UpdateUserAddonsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_name | [string](#string) |  |  |
| addons | [UserAddon](#dashboard-v1alpha1-UserAddon) | repeated |  |






<a name="dashboard-v1alpha1-UpdateUserAddonsResponse"></a>

### UpdateUserAddonsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| user | [User](#dashboard-v1alpha1-User) |  |  |






<a name="dashboard-v1alpha1-UpdateUserDisplayNameRequest"></a>

### UpdateUserDisplayNameRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_name | [string](#string) |  |  |
| display_name | [string](#string) |  |  |






<a name="dashboard-v1alpha1-UpdateUserDisplayNameResponse"></a>

### UpdateUserDisplayNameResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| user | [User](#dashboard-v1alpha1-User) |  |  |






<a name="dashboard-v1alpha1-UpdateUserPasswordRequest"></a>

### UpdateUserPasswordRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_name | [string](#string) |  |  |
| current_password | [string](#string) |  |  |
| new_password | [string](#string) |  |  |






<a name="dashboard-v1alpha1-UpdateUserPasswordResponse"></a>

### UpdateUserPasswordResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |






<a name="dashboard-v1alpha1-UpdateUserRoleRequest"></a>

### UpdateUserRoleRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_name | [string](#string) |  |  |
| roles | [string](#string) | repeated |  |






<a name="dashboard-v1alpha1-UpdateUserRoleResponse"></a>

### UpdateUserRoleResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| user | [User](#dashboard-v1alpha1-User) |  |  |





 

 

 


<a name="dashboard-v1alpha1-UserService"></a>

### UserService


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| DeleteUser | [DeleteUserRequest](#dashboard-v1alpha1-DeleteUserRequest) | [DeleteUserResponse](#dashboard-v1alpha1-DeleteUserResponse) | Delete user by ID |
| GetUser | [GetUserRequest](#dashboard-v1alpha1-GetUserRequest) | [GetUserResponse](#dashboard-v1alpha1-GetUserResponse) | Returns a single User model |
| GetUsers | [.google.protobuf.Empty](#google-protobuf-Empty) | [GetUsersResponse](#dashboard-v1alpha1-GetUsersResponse) | Returns an array of User model |
| CreateUser | [CreateUserRequest](#dashboard-v1alpha1-CreateUserRequest) | [CreateUserResponse](#dashboard-v1alpha1-CreateUserResponse) | Create a new User |
| UpdateUserDisplayName | [UpdateUserDisplayNameRequest](#dashboard-v1alpha1-UpdateUserDisplayNameRequest) | [UpdateUserDisplayNameResponse](#dashboard-v1alpha1-UpdateUserDisplayNameResponse) | Update user display name |
| UpdateUserPassword | [UpdateUserPasswordRequest](#dashboard-v1alpha1-UpdateUserPasswordRequest) | [UpdateUserPasswordResponse](#dashboard-v1alpha1-UpdateUserPasswordResponse) | Update a single User password |
| UpdateUserRole | [UpdateUserRoleRequest](#dashboard-v1alpha1-UpdateUserRoleRequest) | [UpdateUserRoleResponse](#dashboard-v1alpha1-UpdateUserRoleResponse) | Update a single User role |
| UpdateUserAddons | [UpdateUserAddonsRequest](#dashboard-v1alpha1-UpdateUserAddonsRequest) | [UpdateUserAddonsResponse](#dashboard-v1alpha1-UpdateUserAddonsResponse) | Update a single User role |

 



<a name="dashboard_v1alpha1_workspace-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## dashboard/v1alpha1/workspace.proto



<a name="dashboard-v1alpha1-NetworkRule"></a>

### NetworkRule



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| port_number | [int32](#int32) |  |  |
| custom_host_prefix | [string](#string) |  |  |
| http_path | [string](#string) |  |  |
| url | [string](#string) |  |  |
| public | [bool](#bool) |  |  |






<a name="dashboard-v1alpha1-Workspace"></a>

### Workspace



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| owner_name | [string](#string) |  |  |
| spec | [WorkspaceSpec](#dashboard-v1alpha1-WorkspaceSpec) |  |  |
| status | [WorkspaceStatus](#dashboard-v1alpha1-WorkspaceStatus) |  |  |






<a name="dashboard-v1alpha1-WorkspaceSpec"></a>

### WorkspaceSpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| template | [string](#string) |  |  |
| replicas | [int64](#int64) |  |  |
| vars | [WorkspaceSpec.VarsEntry](#dashboard-v1alpha1-WorkspaceSpec-VarsEntry) | repeated |  |
| network | [NetworkRule](#dashboard-v1alpha1-NetworkRule) | repeated |  |






<a name="dashboard-v1alpha1-WorkspaceSpec-VarsEntry"></a>

### WorkspaceSpec.VarsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="dashboard-v1alpha1-WorkspaceStatus"></a>

### WorkspaceStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| phase | [string](#string) |  |  |
| main_url | [string](#string) |  |  |





 

 

 

 



<a name="dashboard_v1alpha1_workspace_service-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## dashboard/v1alpha1/workspace_service.proto



<a name="dashboard-v1alpha1-CreateWorkspaceRequest"></a>

### CreateWorkspaceRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_name | [string](#string) |  | user name |
| ws_name | [string](#string) |  | workspace name |
| template | [string](#string) |  |  |
| vars | [CreateWorkspaceRequest.VarsEntry](#dashboard-v1alpha1-CreateWorkspaceRequest-VarsEntry) | repeated |  |






<a name="dashboard-v1alpha1-CreateWorkspaceRequest-VarsEntry"></a>

### CreateWorkspaceRequest.VarsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="dashboard-v1alpha1-CreateWorkspaceResponse"></a>

### CreateWorkspaceResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| workspace | [Workspace](#dashboard-v1alpha1-Workspace) |  |  |






<a name="dashboard-v1alpha1-DeleteNetworkRuleRequest"></a>

### DeleteNetworkRuleRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_name | [string](#string) |  | user name |
| ws_name | [string](#string) |  | workspace name |
| index | [int32](#int32) |  | network rule index to delete |






<a name="dashboard-v1alpha1-DeleteNetworkRuleResponse"></a>

### DeleteNetworkRuleResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| network_rule | [NetworkRule](#dashboard-v1alpha1-NetworkRule) |  |  |






<a name="dashboard-v1alpha1-DeleteWorkspaceRequest"></a>

### DeleteWorkspaceRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_name | [string](#string) |  | user name |
| ws_name | [string](#string) |  | workspace name |






<a name="dashboard-v1alpha1-DeleteWorkspaceResponse"></a>

### DeleteWorkspaceResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| workspace | [Workspace](#dashboard-v1alpha1-Workspace) |  |  |






<a name="dashboard-v1alpha1-GetWorkspaceRequest"></a>

### GetWorkspaceRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_name | [string](#string) |  | user name |
| ws_name | [string](#string) |  | workspace name |






<a name="dashboard-v1alpha1-GetWorkspaceResponse"></a>

### GetWorkspaceResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workspace | [Workspace](#dashboard-v1alpha1-Workspace) |  |  |






<a name="dashboard-v1alpha1-GetWorkspacesRequest"></a>

### GetWorkspacesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_name | [string](#string) |  | user name |






<a name="dashboard-v1alpha1-GetWorkspacesResponse"></a>

### GetWorkspacesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| items | [Workspace](#dashboard-v1alpha1-Workspace) | repeated |  |






<a name="dashboard-v1alpha1-UpdateWorkspaceRequest"></a>

### UpdateWorkspaceRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_name | [string](#string) |  | user name |
| ws_name | [string](#string) |  | workspace name |
| replicas | [int64](#int64) | optional | workspace to update |






<a name="dashboard-v1alpha1-UpdateWorkspaceResponse"></a>

### UpdateWorkspaceResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| workspace | [Workspace](#dashboard-v1alpha1-Workspace) |  |  |






<a name="dashboard-v1alpha1-UpsertNetworkRuleRequest"></a>

### UpsertNetworkRuleRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_name | [string](#string) |  | user name |
| ws_name | [string](#string) |  | workspace name |
| network_rule | [NetworkRule](#dashboard-v1alpha1-NetworkRule) |  | network rule to upsert |
| index | [int32](#int32) |  | network rule index to update. insert if index is out of length |






<a name="dashboard-v1alpha1-UpsertNetworkRuleResponse"></a>

### UpsertNetworkRuleResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |
| network_rule | [NetworkRule](#dashboard-v1alpha1-NetworkRule) |  |  |





 

 

 


<a name="dashboard-v1alpha1-WorkspaceService"></a>

### WorkspaceService


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateWorkspace | [CreateWorkspaceRequest](#dashboard-v1alpha1-CreateWorkspaceRequest) | [CreateWorkspaceResponse](#dashboard-v1alpha1-CreateWorkspaceResponse) | Create a new Workspace |
| DeleteWorkspace | [DeleteWorkspaceRequest](#dashboard-v1alpha1-DeleteWorkspaceRequest) | [DeleteWorkspaceResponse](#dashboard-v1alpha1-DeleteWorkspaceResponse) | Delete workspace |
| UpdateWorkspace | [UpdateWorkspaceRequest](#dashboard-v1alpha1-UpdateWorkspaceRequest) | [UpdateWorkspaceResponse](#dashboard-v1alpha1-UpdateWorkspaceResponse) | Update workspace |
| GetWorkspace | [GetWorkspaceRequest](#dashboard-v1alpha1-GetWorkspaceRequest) | [GetWorkspaceResponse](#dashboard-v1alpha1-GetWorkspaceResponse) | Returns a single Workspace model |
| GetWorkspaces | [GetWorkspacesRequest](#dashboard-v1alpha1-GetWorkspacesRequest) | [GetWorkspacesResponse](#dashboard-v1alpha1-GetWorkspacesResponse) | Returns an array of Workspace model |
| UpsertNetworkRule | [UpsertNetworkRuleRequest](#dashboard-v1alpha1-UpsertNetworkRuleRequest) | [UpsertNetworkRuleResponse](#dashboard-v1alpha1-UpsertNetworkRuleResponse) | Upsert workspace network rule |
| DeleteNetworkRule | [DeleteNetworkRuleRequest](#dashboard-v1alpha1-DeleteNetworkRuleRequest) | [DeleteNetworkRuleResponse](#dashboard-v1alpha1-DeleteNetworkRuleResponse) | Remove workspace network rule |

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

