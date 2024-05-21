# User

`User` is a cluster-scoped Kubernetes CRD which represents a developer or user who use Workspace.

When you create User, Kubernetes Namespace is created and bound to the User.

```yaml
apiVersion: cosmo-workspace.github.io/v1alpha1
kind: User
metadata:
  name: voldemort
spec:
  displayName: Tom Riddle    # displayName is a free format name shown on Dashboard
  authType: password-secret  # authType is authentication type for this User
  roles:                     # roles is a list of role. role name end with `-admin` is treated as admin for the prefix.
  - name: cosmo-admin        # `cosmo-admin` is a privileged role on Dashboard (administrator of cosmo)
  addons:                    # addons is a list of UserAddon, which is the set of additional resources in one-to-one with User
  - template:
      name: cosmo-username-headers
status:
  namespace:
    apiVersion: v1
    creationTimestamp: "2023-09-12T13:19:42Z"
    kind: Namespace
    name: cosmo-user-voldemort
    uid: f1edc3bb-8de7-47ec-a0bd-64751322b491
  phase: Active
```

## AuthType

Authentication type. Currently supports:

|type|descroption|
|:--|:--|
|`password-secret`| Builtin authentication enabled by default. You can use this type anywhere with no configuration. |
|`ldap`| Use ldap server to authentication. You need to configure ldap server info in installing |

## Role

Role is a role for User. Role itself is just like a label.

If `Template` or `ClusterTemplate` has an annotation `cosmo-workspace.github.io/userroles`, the Template is only shown on the Users who has the Role.

## UserAddon

UserAddon is a set of Kubernetes manifests for each Users, which are required to be created by User.

For example, limitting resource quota, granting ClusterRole, binding external ID, account or IAM, create PersistentVolume for the User, and so on. 

UserAddon can be defined as `Template` or `ClusterTemplate` if including cluster-scoped resources.

See [TEMPLATE-ENGINE.md](https://github.com/cosmo-workspace/cosmo/blob/main/docs/TEMPLATE-ENGINE.md) for deepdive into `Template` and `ClusterTemplate`.

<details>
<summary>resource-limitter.yaml</summary>

```yaml
apiVersion: cosmo-workspace.github.io/v1alpha1
kind: Template
metadata:
  annotations:
    cosmo-workspace.github.io/disable-nameprefix: "true"
  labels:
    cosmo-workspace.github.io/type: useraddon
  name: resource-limitter
spec:
  description: "limit user resources"
  rawYaml: |
    apiVersion: v1
    kind: ResourceQuota
    metadata:
      name: quota
      namespace: '{{NAMESPACE}}'
    spec:
      hard:
        limits.cpu: '{{LIMIT_TOTAL_CPU_CORE}}'
        limits.memory: '{{LIMIT_TOTAL_MEM_GB}}Gi'
        {{STORAGE_CLASS_NAME}}.storageclass.storage.k8s.io/requests.storage: '{{LIMIT_TOTAL_VOLUME_GB}}Gi'
  requiredVars:
    - var: LIMIT_TOTAL_CPU_CORE
      default: "2"
    - var: LIMIT_TOTAL_MEM_GB
      default: "16"
    - var: LIMIT_TOTAL_VOLUME_GB
      default: "64"
    - var: STORAGE_CLASS_NAME
      default: "gp2"
```

</details><br>

UserAddon can be generated via `cosmoctl tmpl gen` command, same as WorkspaceTemplate.

### Create UserAddon

1.  Prepare the set of Kubernetes YAML by your own.

    All you have to do is to prepare your own Kubernetes YAMLs that is deployable.

    For example, you prepare like here.

    ```sh
    $ ls /tmp/useraddon-example
    kustomization.yaml   cluster-role.yaml   cluster-rolebinding.yaml   sa.yaml
    ```

    <details>
    <summary>kustomization.yaml</summary>

    ```yaml
    apiVersion: kustomize.config.k8s.io/v1beta1
    kind: Kustomization

    resources:
    - cluster-role.yaml
    - cluster-rolebinding.yaml
    - sa.yaml
    ```

    </details><br>

    <details>
    <summary>cluster-role.yaml</summary>

    ```yaml
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRole
    metadata:
      name: cluster-view
    rules:
    - apiGroups:
      - "*"
      resources:
      - "*"
      verbs:
      - get
      - list
      - watch
    ```

    </details><br>

    <details>
    <summary>cluster-rolebinding.yaml</summary>

    ```yaml
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRoleBinding
    metadata:
      name: cluster-admin
    roleRef:
      apiGroup: rbac.authorization.k8s.io
      kind: ClusterRole
      name: cluster-view
    subjects:
    - kind: ServiceAccount
      name: code-server
    ```

    </details><br>

    <details>
    <summary>sa.yaml</summary>

    ```yaml
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: code-server
    ```
    </details><br>

    Be sure it can be deployed

    ```sh
    kustomize build . | kubectl apply -f - --dry-run=server
    ```

    > ✅Note:
    >
    > Be sure it can be deployed as a single WebIDE deployment without any networking or storage troubles!
    >
    > Delete them if you applied the manifest actually. `kustomize build . | kubectl delete -f -`

2.  Generate WorkspaceTemplate

    Pass kustomize-generated manifest to `cosmoctl tmpl gen` command by stdin.

    ```sh
    kustomize build . | cosmoctl tmpl gen --cluster-scope --useraddon -o addon.yaml
    ```

    > ✅Note: 
    >
    > Be sure to execute with `--useraddon` option.
    >
    > `--cluster-scope` is only required when including cluster-scoped resources like ClusterRole.
    
In order to create a default user addon, which is applied to all Users automatically, annotate `useraddon.cosmo-workspace.github.io/default: "true"` on the Template.

## Annotations

UserAddons with the following annotations have special behavior.

| Annotatio keys | Avairable values(default) | Description | cosmoctl option |
|:--|:--|:--|:--|
| `useraddon.cosmo-workspace.github.io/default` | `["true", "false"]`("false") | UserAddon with this annotation is applied to all Users automatically | `--useraddon-set-default` |
| `cosmo-workspace.github.io/disable-nameprefix` | `["true", "false"]`("false") | UserAddon with this annotation is applied to all Users automatically | `--disable-nameprefix` |
| `cosmo-workspace.github.io/userroles` | comma-separated UserRoles(None) | User who use this Template must have all of the UserRoles specified in this annotation | `--userroles` |
| `cosmo-workspace.github.io/required-useraddons` | comma-separated UserAddon names(None)  | User who use this Template must be attached all of the UserAddons specified in this annotation | `--required-useraddons` |


### More infomation

When you create `Workspace`, you can also see the Kubernetes resource `Instance` is created.

