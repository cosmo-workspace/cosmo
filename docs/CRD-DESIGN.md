# Custom Resource Definition Design

Kubernetes has extension feature called [Custom Resource Definition(CRD)](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)

COSMO has 3 CRD resources.
- `Workspace`: Abstract resource for WebIDE
- `Template` and `Instance`: Core resources for COSMO Template Engine, 

COSMO is a manager of WebIDE, but `Templates` and `Instances` are designed to be generic and can be used not only for the WebIDE containers but others.

It is called COSMO Template engine, designed for a replication of s the set of Kubernetes manifests in a cluster.

COSMO Template engine is picking the best of both overlay-based [`Kustomize`](https://github.com/kubernetes-sigs/kustomize) and variable-based [`Helm`](https://helm.sh).

## Template

`Template` is a collection of standard Kubernetes YAML manifests.

The following example is a `nginx` template with Ingress, Service, and Deployment resources.

```yaml
apiVersion: cosmo.cosmo-workspace.github.io/v1alpha1
kind: Template
metadata:
  name: nginx
spec:
  requiredVars:
    - var: '{{IMAGE_TAG}}'
      default: latest
    - var: '{{DOMAIN}}'
  rawYaml: |
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      labels:
        cosmo/instance: '{{INSTANCE}}'
        cosmo/template: nginx
      name: nginx
    spec:
      rules:
      - host: '{{INSTANCE}}-{{NAMESPACE}}.{{DOMAIN}}'
        http:
          paths:
          - path:
            pathType: Prefix
            backend:
              service:
                name: '{{INSTANCE}}-nginx'
                port: 
                  number: 80
    ---
    apiVersion: v1
    kind: Service
    metadata:
      labels:
        cosmo/instance: '{{INSTANCE}}'
        cosmo/template: nginx
      name: nginx
      namespace: '{{NAMESPACE}}'
    spec:
      ports:
      - name: main
        port: 80
        protocol: TCP
      selector:
        cosmo/instance: '{{INSTANCE}}'
        cosmo/template: nginx
      type: ClusterIP
    ---
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      labels:
        cosmo/instance: '{{INSTANCE}}'
        cosmo/template: nginx
      name: nginx
      namespace: '{{NAMESPACE}}'
    spec:
      replicas: 1
      selector:
        matchLabels:
          cosmo/instance: '{{INSTANCE}}'
          cosmo/template: nginx
      template:
        metadata:
          labels:
            cosmo/instance: '{{INSTANCE}}'
            cosmo/template: nginx
        spec:
          containers:
          - image: 'nginx:{{IMAGE_TAG}}'
            name: nginx
            ports:
            - containerPort: 80
              name: main
              protocol: TCP
```

### Template Variables
Template has variables. It is a minimal feature for text-based YAML replacement, so that you do not need to understand the detailed syntax.

Variables are defined as UPPERCASE letters + underscores (A-Z_) enclosed in `{{` and `}}`.
And it will be replaced when a `Instance` is created from the Template. 

There are 2 types of variables, pre-defined variables and user-defined variables.

The pre-defined variables are only as follows.

| Variables     | Description                           |
|:--------------|:--------------------------------------|
| {{INSTANCE}}  | Instance name                         |
| {{NAMESPACE}} | Namespace name in which instance runs |

In the example Template, `{{IMAGE_TAG}}` and `{{DOMAIN}}` are user-defined variables.

A user-defined variable in `requiredVars` requires the value of the variable when creating `Instance` by Kubernetes's Validating Webhook.

### Resource Name
The each resource name (.metadata.name) will be prefixed with `{{INSTANCE}}-` even without using template variables. 
In other words, the resource name will always be `'{{INSTANCE}}-nginx'`.

> Note:
> In the example, the backend service name of Ingress has prefix `{{INSTANCE}}-`.
> You need to explicitly add `{INSTANCE}}-` if you need to point to another resource in the Template. 

## Instance

`Instance` is a collection of Kubernetes resource entities based on a `Template`.

When `Instance` is created, the Kubernetes resources defined in `Template` will be created.

`Instance` is the owner resource of each created Kubernetes resources. 
So when you delete `Instance`, all the child resources will be deleted.

An example of Instance using `nginx` Template.

```yaml
apiVersion: cosmo.cosmo-workspace.github.io/v1alpha1
kind: Instance
metadata:
  name: my-nginx1
  namespace: default
spec:
  template:
    name: nginx
  vars:
    "{{IMAGE_TAG}}": alpine
    "{{DOMAIN}}": example.cosmo-workspace.github.io
  override:
    scale:
      target:
        apiVersion: apps/v1
        kind: Deployment
        name: nginx
      replicas: 1
    network:
      ingress:
        targetName: nginx
        annotations:
          cosmo/sample: sample-annotation
        rules:
        - http:
            paths:
            - path: /add
              pathType: Exact
              backend:
                service:
                  name: nginx
                  port:
                    number: 9090
      service:
        targetName: nginx
        ports:
        - name: add
          port: 9090
          protocol: TCP
          targetPort: 9090
    patchesJson6902:
    - target:
        apiVersion: v1
        kind: Service
        name: nginx
      patch: |
        [
          {
            "op": "replace",
            "path": "/spec/type",
            "value": "LoadBalancer"
          }
        ]
```

### Template selector
Each resources will be created based on `nginx` Template with a OwnerReference of the Instance . 
By deleting the Instance, all resources created by Instance will be automatically deleted by OwnerReference.

If the Template is changed, it will be dynamically applied to the running resources.

### Vars
`vars` is a key-value Map of the user-defined variables in the `requiredVars` of the Template.

Pre-defined Template variables and user-defined variables defined in `vars` will be replaced before applying to Kubernetes.

### Override
Override is one of the most important features of Instance, which supports dynamic change of Template.

Currently, the following Overrides are natively supported.
- Start and stop Pods by overriding `.spec.replicas` in Workload resources (Deployment etc.)
- Override `ServicePort` of Service.
- Override `IngressRule` for Ingress.
- Override `Annotations` for Ingress.

Also you can patch with `patchesJson6902` in [RFC6902](https://www.rfc-editor.org/rfc/rfc6902.html) format for non-natively supported override.

Resource names will be recognized as an Instance name prefixed.
Therefore you don't need to prefix `my-nginx1-` on resource names, like Deployment name in override targets.

## Workspace
`Workspace` is a wrapper resource of `Instance`. It is designed for abstraction of developer WebIDE containers.

When you create `Workspace`, COSMO automatically creates `Instance` and then actual Workspace Pods or Services will be created, which is defined in the specified `Template`.

`Workspace` has the following features for the WebIDE container.

- Easy to stop or start Workspace.
- Abstraction of Networking spec.
- Dynamic opening or closing ports for dev servers running in WebIDE.
- Auto authentication to the WebIDE container and all opened ports.

An example of Workspace.

```yaml
apiVersion: workspace.cosmo-workspace.github.io/v1alpha1
kind: Workspace
metadata:
  name: my-workspace
  namespace: cosmo-user-tom
spec:
  replicas: 1
  template:
    name: example-cs
  vars:
    COSMO_AUTH_PROXY_IMAGE: v0.17.0-beta2
  network:
  - group: main
    httpPath: /
    portName: http
    portNumber: 8080
```

### Template, Vars
These will be propagated to the `Instance` specification.

However, only `Workspace Template` can be specified in Template spec. It is just a Template but labeled with `cosmo/type: workspace` and some annotations.

See [Workspace.md](Workspace.md)　for the detail of `Workspace Template`.

### Replicas
Just a number of replica of Pod. 1 mean Workspace is running and 0 is stopped.
It will be propagated to the `Instance scaling override`.

### Network
Network is a abstract definition of networking features.
It will be propagated to the `Instance network override` as Service and Ingress override.

See [Workspace.md](Workspace.md)　for the detail.
