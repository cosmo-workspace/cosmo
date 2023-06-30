# Getting Started

## 1. Requirement

### ✓ Kubernetes Version
  
Version ≥ 1.19

COSMO is targeting to run on any Kubernetes distribution.

### 🌐 Networking & DNS

Networking is vary dependent on your environment.

For all users to securely connect to the workspaces, you should use TLS. 
So we recommend that you configure the following in your DNS Name server.
However, since it costs to acquire a domain, this Getting Started will also describe without DNS.

Replace `cosmo.example.com` to your domain or subdomain for COSMO.
- `*.cosmo.example.com` to Ingress or LoadBalancer for Workspaces.
- `cosmo.example.com` to Ingress or LoadBalancer for COSMO Dashboard.

COSMO supporting the following networking feature in Kubernetes.
- ✅ [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/)
- ✅ [LoadBalancer Service](https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer)
- ✅ [NodePort Service](https://kubernetes.io/docs/concepts/services-networking/service/#nodeport)

The followings are not supported.
- ❌ [Gateway API](https://gateway-api.sigs.k8s.io/)
- ❌ Service mesh or other networking extensions (Istio etc.)

For each networking feature, accessibility from your browser to load balancers or Kuberetes Nodes is required.

- **For Ingress (Recommended)**:

  COSMO uses a Host-header routing for each Workspace.
  COSMO does not require a specific Ingress Controller but can be used with any Ingress Controller.

  Configure a wildcard-host domain record in your name server targeting to your Load Balancer created by your Ingress Controllers.

  For example in Amazon EKS using AWS Load Balancer Controller and Route53 for a name server

- **For LoadBalancer Service**:

  COSMO uses a DNS name or IP address automatically created by LoadBalancer Service.

  We recommend configure canonical name or alias name to load balancer's address in your name server so that you can use TLS.

  However if you can access the auto-created Load Balancer with insecure access, you requires no more networking configurations.

  For example:
  ```shell
  $ kubectl create svc loadbalancer example --tcp=80:8080
  service/example created

  $ kubectl get svc example
  NAME      TYPE           CLUSTER-IP       EXTERNAL-IP                                                                   PORT(S)        AGE
  example   LoadBalancer   10.100.202.219   a9eea00d47b484d81a9285d009b23a3b-496716172.ap-northeast-1.elb.amazonaws.com   80:30845/TCP   72s
  ```

  Check if you can resolve the DNS name in `EXTERNAL-IP` and access Dashboard from your browser.
  (Be sure to delete the example service not to cost you too much.)

  Or you should configure the DNS record to `EXTERNAL-IP` in your name server manually.

- **For NodePort Service**:

  Access to your Kubernetes Nodes by a representative DNS name as a single endpoint is required.

  We recommend to install a tool that syncs Node IP addresses with name server's specific record. Such as https://github.com/calebdoxsey/kubernetes-cloudflare-sync, 
  https://github.com/jlandowner/kubernetes-route53-sync and so on.

### 💾 Volumes

To persist workspace data, each workspace templates should include a [Persistent Volume Claim](https://kubernetes.io/docs/concepts/storage/persistent-volumes/).

You need to configure [Dynamic Provisioning](https://kubernetes.io/docs/concepts/storage/dynamic-provisioning/) in Kubernetes, which does not require a Persistent Volume manually created by administrator.

- **On-prem or self-installed cluster**

  See [Storage Class](https://kubernetes.io/docs/concepts/storage/storage-classes/) or install storage extentions suporting Dynamic Provisioning such as [Longhorn](https://longhorn.io/docs/1.2.0/volumes-and-nodes/create-volumes/) or [OpenEBS](https://github.com/openebs/openebs).

- **Cloud providers's cluster**

  See the official docs. Almost all cloud providers supports their own storage driver.

  For example:
  - [Amazon EKS](https://docs.aws.amazon.com/eks/latest/userguide/storage.html)

  - [Google Kubernetes Engine](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/ssd-pd)

  - [Azure Kubernetes Service](https://docs.microsoft.com/azure/aks/concepts-storage#storage-classes)

## 2. Install COSMO on your Kubernetes Cluster

Confirm followings before installation.
- [kubectl](https://kubernetes.io/ja/docs/tasks/tools/install-kubectl/) and [helm](https://helm.sh/docs/intro/install/) is installed in PATH.
- kubectl current context to the cluster where you install COSMO.

### 2-1. Install [`cert-manager`](https://cert-manager.io/)

cosmo requires a lot of tls certificates.

The easiest way to prepare certificates is to use [`cert-manager`](https://cert-manager.io/) (>=v1.0.0)

See the offical install docs https://cert-manager.io/docs/installation/

```sh
kubectl apply -f https://github.com/jetstack/cert-manager/releases/latest/download/cert-manager.yaml
```

### 2-2. Install [`COSMO Controller Manager`](https://github.com/cosmo-workspace/cosmo/pkgs/container/cosmo-controller-manager)

Add helm repo

```sh
helm repo add cosmo https://cosmo-workspace.github.io/charts
```

Install cosmo-controller-manager

```sh
helm upgrade --install -n cosmo-system --create-namespace cosmo-controller-manager cosmo/cosmo-controller-manager
```

Install default Templates.

```sh
LATEST_TAG=`curl https://api.github.com/repos/cosmo-workspace/cosmo/releases/latest | jq  -r '.tag_name'`

```

## Install [`COSMO Dashboard`](https://github.com/cosmo-workspace/cosmo/pkgs/container/cosmo-dashboard)

Since you need to access Dashboard from a browser, install Dashboard with the appropriate network options.

- **For Ingress**:

  Prepare COSMO Dashboard DNS name like `cosmo.example.com`.

  Create `values.yaml` to configure Ingress.
  For example:

  ```yaml
  ingress:
    enabled: true
    ## Uncomment if you use not default IngressClass
    # className: YOUR_INGRESS_CLASS_NAME
    ## Uncomment and set key: value if you append annotation for Ingress
    # annotations:
    hosts:
    - host: cosmo.example.com # Replace to yours
      paths:
      - path: /
        pathType: Prefix
  cert:
    dnsName: cosmo.example.com
  ```

  >📘 Note:
  >This example uses a self-signed certificate with hostname `cosmo.example.com`.
  >You can configure your own certificate. See [charts repository](https://github.com/cosmo-workspace/charts/blob/main/README.md#install-options) for details.

  Install Dashobard with the `values.yaml`

  ```sh
  helm upgrade --install -n cosmo-system cosmo-dashboard cosmo/cosmo-dashboard -f values.yaml
  ```

- **For LoadBalancer Service**:

  If you access Dashboard with custom hostname, prepare COSMO Dashboard DNS name like `cosmo.example.com` and install.

  ```sh
  helm upgrade --install -n cosmo-system cosmo-dashboard cosmo/cosmo-dashboard --set service.type=LoadBalancer --set dnsName=cosmo.example.com
  ```

  Or install without dnsName.

  ```sh
  helm upgrade --install -n cosmo-system cosmo-dashboard cosmo/cosmo-dashboard --set service.type=LoadBalancer
  ```

  Check LoadBalancer service status.

  ```sh
  kubectl get service
  ```

  Example output

  ```
  NAME      TYPE           CLUSTER-IP       EXTERNAL-IP                                                                   PORT(S)        AGE
  example   LoadBalancer   10.100.202.219   a9eea00d47b484d81a9285d009b23a3b-496716172.ap-northeast-1.elb.amazonaws.com   80:30845/TCP   72s
  ```

  Next, if you access Dashboard with custom hostname, configure DNS record to `EXTERNAL-IP` in your name server.

  Or create a self-signed certificate for `EXTERNAL-IP` hostname via cert-manager.
  (If `EXTERNAL-IP` is actual IP address and not hostname, you cannot use a self-signed certificate.)

  ```sh
  kubectl edit cert cosmo-dashboard-cert -n cosmo-system
  ```

  Add `EXTERNAL-IP` in `spec.dnsNames`.

  ```
  spec:
    dnsNames:
    - cosmo-dashboard.cosmo-system.svc
    - cosmo-dashboard.cosmo-system.svc.cluster.local
    - a9eea00d47b484d81a9285d009b23a3b-496716172.ap-northeast-1.elb.amazonaws.com # Add EXTERNAL-IP hostname
  ```

  Restart COSMO Dashboard to reload the certificate.

  ```sh
  kubectl rollout restart deploy/cosmo-dashboard -n cosmo-system
  ```

- **For NodePort Service**:

  Replace `node.example.com` to your node representative DNS name and install.

  ```sh
  helm upgrade --install -n cosmo-system cosmo-dashboard cosmo/cosmo-dashboard \
  --set service.type=NodePort \
  --set service.nodePort=32080 \
  --set dnsName=node.example.com
  ```

  >📘 Note:
  >This example uses a self-signed certificate with hostname `node.example.com`.
  >You can configure your own certificate. See [charts repository](https://github.com/cosmo-workspace/charts/blob/main/README.md#install-options) for details.


Now access COSMO Dashboard URL you configured from your browser. 
You may see a certificate alert for self-signed certificate but continune.

> 💡 See [charts repository](https://github.com/cosmo-workspace/charts/blob/main/README.md#install-options) for more information on helm installation options.

> If your browser deny the access to the invalid certificate URL, install CA certificate on your browser or PC.
> 
> Download the CA certficate.
> 
> ```sh
> kubectl get secret -n cosmo-system dashboard-server-cert -o jsonpath='{.data.ca\.crt}' | base64 -d > ca.crt
> ```
> 
> Install `ca.crt` to your browser or PC.

### 2-3. Install cosmoctl

Download binary from [latest release](https://github.com/cosmo-workspace/cosmo/releases/latest) and extract it into PATH.

## 3. Create first User

Use cosmoctl to create first User.

```sh
cosmoctl user create admin --admin
```

Output:

```sh
Successfully created user admin
Default password: DEFAULT_PASSWORD
```

Access the dashboard with a browser and login as admin with the output password. 

> ⚠️Caution:
>
> You will be required to change password but DO NOT enter the password used in other services.
> You are now not using TLS and insecure!

Now you are ready to use COSMO 🚀

## 4. Create example Template

However there is still NO Template.
Create example code-server Template from the [official helm chart](https://github.com/cdr/code-server/tree/main/ci/helm-chart).

### 4-1. Clone the code-server project.

```sh
git clone https://github.com/cdr/code-server.git
```

### 4-2. Configure code-server template for your environment

Configure code-server helm's `values.yaml` for your environmet. Probably it will be about networking.

> ⚠️Caution:
> This example not configure TLS setting, but you should use TLS in your actual Workspace.

- **For Ingress**:

  Create `your-values.yaml`.
  For example:

  ```sh
  cat <<EOF > your-values.yaml
  # Replace ".cosmo.example.com" to your domain. 
  # But DO NOT edit "{{INSTANCE}}" and "{{NAMESPACE}}". These valiables will be automatically replaced.
  ingress:
    enabled: true
    hosts:
      - host: "http-{{INSTANCE}}-{{NAMESPACE}}.cosmo.example.com"
        paths:
          - /

  # Use COSMO Auth Proxy instead of code-server's default authorization.
  extraArgs:
  - --auth=none

  # Use COSMO user's serviceaccount
  serviceAccount:
    create: false
    name: default
  EOF
  ```

  Execute helm to generate k8s configs.

  ```sh
  helm template code-server ./ci/helm-chart/ -f your-values.yaml --skip-tests > your-code-server.yaml
  ```

  Replace `.cosmo.example.com` to your domain and execute cosmoctl to generate COSMO Template.

  ```sh
  cat your-code-server.yaml | cosmoctl tmpl generate --name=example-cs --workspace \
    --workspace-urlbase='http://{{NETRULE_GROUP}}-{{INSTANCE}}-{{NAMESPACE}}.cosmo.example.com' > cosmo-template.yaml
  ```

  >⚠️Caution:
  >DO NOT edit `{{NETRULE_GROUP}}`, `{{INSTANCE}}` and `{{NAMESPACE}}`.
  >These valiables will be automatically replaced.


- **For LoadBalancer Service**:

  Create `your-values.yaml`.
  For example:

  ```sh
  cat <<EOF > your-values.yaml
  # Service type
  service:
    type: LoadBalancer

  # Use COSMO Auth Proxy instead of code-server's default authorization.
  extraArgs:
  - --auth=none

  # Use COSMO user's serviceaccount
  serviceAccount:
    create: false
    name: default
  EOF
  ```

  Execute helm to generate k8s configs.

  ```sh
  helm template code-server ./ci/helm-chart/ -f your-values.yaml --skip-tests > your-code-server.yaml
  ```

  Execute cosmoctl to generate COSMO Template.
  
  ```sh
  cat your-code-server.yaml | cosmoctl tmpl generate --name=example-cs --workspace \
    --workspace-urlbase='http://{{LOAD_BALANCER}}:{{PORT_NUMBER}}' > cosmo-template.yaml
  ```

  >⚠️Caution:
  >DO NOT edit `{{LOAD_BALANCER}}` and `{{PORT_NUMBER}}`.
  >These valiables will be automatically replaced.


- **For NodePort Service**:

  Create `your-values.yaml`.
  For example:

  ```sh
  cat <<EOF > your-values.yaml
  # Service type
  service:
    type: NodePort

  # Use COSMO Auth Proxy instead of code-server's default authorization.
  extraArgs:
  - --auth=none

  # Use COSMO user's serviceaccount
  serviceAccount:
    create: false
    name: default
  EOF
  ```

  ```sh
  helm template code-server ./ci/helm-chart/ -f your-values.yaml --skip-tests > your-code-server.yaml
  ```

  Replace `node.example.com` to your node representative DNS name.

  ```sh
  cat your-code-server.yaml | cosmoctl tmpl generate --name=example-cs --workspace \
    --workspace-urlbase='http://node.example.com:{{NODEPORT_NUMBER}}' > cosmo-template.yaml
  ```

  >⚠️Caution:
  >DO NOT edit `{{NODEPORT_NUMBER}}`.
  >This valiables will be automatically replaced.

### 4-3. Apply Template

```
kubectl create -f cosmo-template.yaml
```

Now you are ready to use Template `example-code-server` 🙌

## 5. Create example Workspace

From browser, create a example workspace with created Template.


<video src="https://user-images.githubusercontent.com/9918931/136198651-75626165-d421-4fa8-bbf0-932d7020ae84.mp4"></video>


## 6. Cleanup

Uninstall COSMO Dashboard.

```sh
helm uninstall -n cosmo-system cosmo-dashboard
```

Uninstall COSMO Controller Manager and CRD.

This will remove all Workspaces and Templates.

```sh
helm uninstall -n cosmo-system cosmo-controller-manager
```
