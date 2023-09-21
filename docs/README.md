# COSMO - Open source WebIDE & DevEnvironment Platform on Kubernetes

<img src="https://raw.githubusercontent.com/cosmo-workspace/cosmo/main/logo/logo-with-name-small.png">

[![kubernetes](https://img.shields.io/badge/kubernetes-grey.svg?logo=Kubernetes)](https://kubernetes.io/)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/cosmo-workspace/cosmo/blob/main/LICENSE)
[![GoReportCard](https://img.shields.io/badge/go%20report-A+-brightgreen.svg?style=flat)](https://goreportcard.com/report/github.com/cosmo-workspace/cosmo)
[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/cosmo)](https://artifacthub.io/packages/search?repo=cosmo)

An Open Source WebIDE & Cloud DevEnvironment Platform on Kubernetes, like GitHub CodeSpaces or AWS Cloud9.

COSMO manages the WebIDE as a container. Integrating with the great ecosystem of Kubernetes, you can create a robust, scalable, flexible private cloud development environment.

# Feature

- Flexible: Any WebIDE as you like
- Minimal: A small set of kubernetes-native components to manage WebIDE containers and integrate with kubernetes ecosystem.
- Authentication: User authentication for each Workspaces
- Dynamic dev-server ports: Open dev-server port on WebIDE container and expose them with authentication

<video src="https://github-production-user-asset-6210df.s3.amazonaws.com/48989258/265190584-8bfcbede-36bd-47be-a1e7-04e4d6326521.mp4"></video>

## Any WebIDE as you like

Existing cloud devenvironment services such as GitHub CodeSpaces and AWS Cloud9 are easy to try. However they force developers to use unfamilier WebIDE bundled with the service.

In addition, "Node.js development template" is never the single template in real projects. Different projects require different packages and libraries or different version and tool combinations.

COSMO does not bundle WebIDE itself but BROI, Bring Your Own WebIDE Container Image.

The easiest way is to use a pre-built WebIDE container image provided by great open soruce projects like: 
- https://hub.docker.com/r/codercom/code-server
- https://hub.docker.com/r/gitpod/openvscode-server
- https://hub.docker.com/r/linuxserver/openvscode-server
- https://jupyter-docker-stacks.readthedocs.io/en/latest/

We recommend to build your own container image base on the above image with the programming languages, extensions, development tools, etc. which required for your project, in order to start developing your project ASAP by just launching Workspace.

## Minimal set of managing WebIDE containers
COSMO is a minimal set of components, Kubernetes CRD, Controller and Dashboard.

There is No database and all states are stored on Kubernetes.

Integrating with the great ecosystem of Kubernetes, you can run a robust, scalable, flexible cloud development environment.
- Autoscaling: [Cluster AutoScaler](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler) / [Karpenter](https://karpenter.sh/)
- Storage: [CSI Drivers](https://kubernetes-csi.github.io/docs/drivers.html)
- Backup: [velero](https://velero.io/) / [CSI Snapshot](https://kubernetes-csi.github.io/docs/snapshot-restore-feature.html)
- Limiting compute resources: [ResourceQuota](https://kubernetes.io/docs/concepts/policy/resource-quotas/)
- CodeRepository: Self-hosted Git server / AWS CodeCommit / Cloud Source Repositories

## Authentication
COSMO bundles a traefik authentication plugin and you can protect Workspace URLs by default.

## Dynamic dev-server ports
You can expose dynamic dev-server ports in WebIDE container.
Automatically configure routing and secure URL with User authentication.


# Architecture overview

![architecture](assets/architecture.drawio.svg)

- Dashboard: User Interface for managing COSMO resources.
- Controller: Sync Instances of Template continually by applying Kubernetes manifests in the Template.
- Traefik & AuthPlugin: [Traefik](https://traefik.io/) is router of Workspaces and [Traefik](https://traefik.io/) plugin does User authentication to protect Workspaces.

COSMO has 3 main concepts, which is implemented as Kubernetes CRD.

- `Workspace`: A single WebIDE instance of Template.
- `Template`: A set of Kubernetes manifests required for a WebIDE container to run. For most cases, they are Kubernetes Deployment, Service and PersistentVolumeClaim. Template can include any kind of Kubernetes manifests.
- `User`: An identity of a developer. User can run one or more Workspaces to choose properly Template for developing each apps or projects. In Kubernetes terms, it is a [`Namespace`](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/) actually. You can protect your workspace network traefik by User authentication.
- `UserAddon`: A set of Kubernetes manifests required for each Users. For example, you can include like AWS IAM Role for ServiceAccount, Kubernetes RBAC, PersitentVolume for shared filesystem and so on.

See [CRD-DESIGN.md](https://github.com/cosmo-workspace/cosmo/blob/main/docs/CRD-DESIGN.md) for more details.


# Getting Started

See [GETTING-STARTED.md](https://github.com/cosmo-workspace/cosmo/blob/main/docs/GETTING-STARTED.md)


# Distinguished by other self-hosted products

| Name | Subscription (including free plan) | Database required | Dynamic dev server network | Workspace Authentication | Dynamic Port Authentication | Custom WebIDE Image |
|:---|:---|:---|:---|:---|:---|:---|
| **COSMO** | - | **No database required** | ✅ | ✅ | ✅ | ✅ |
| [`Eclipse Che local install`](https://www.eclipse.org/che/) | - | Yes (Postgres) -> No [from 7.62.0](https://che.eclipseprojects.io/2023/03/20/@ilya.buziuk-decommissioning-postgresql-database.html) | - | [✅](https://www.eclipse.org/che/docs/che-7/administration-guide/authenticating-users/) | - | [✅ devfile](https://eclipse.dev/che/docs/stable/end-user-guide/devfile-introduction/) | 
| [`Coder self-hosted`](https://coder.com/) | [Yes](https://coder.com/pricing) | [Yes (Postgres)](https://coder.com/docs/v2/latest/about/architecture)| [✅](https://coder.com/docs/v2/latest/about/architecture#coderd) | [✅](https://coder.com/docs/v2/latest/admin/auth) | ✅ |
| [`Gitpod Self-Hosted`](https://www.gitpod.io/self-hosted) | [No longer suppoerted](https://www.gitpod.io/docs/configure/self-hosted/latest) | Yes (MySQL) | [✅](https://www.gitpod.io/docs/config-ports) | [✅](https://www.gitpod.io/docs/configure/authentication#authentication) | [✅](https://www.gitpod.io/docs/config-ports) | ✅ |

Existing products are so great and rich features but they requires learning about the products.
[Gitpod choosed that no longer support self-hosted. they says:](https://www.gitpod.io/blog/introducing-gitpod-dedicated)
> Despite all that effort, self-hosted Gitpod has been increasingly difficult for us to support and it has shown to be a burden for our clients to manage and operate their own Gitpod instances. 

While COSMO does not have much features, it is designed to be minimumn and remove feature that can be use Kubernetes ecosystem. And that is easy for standard Kubernetes administrators to operate with only knowledge of Kubernetes, as it only requires a set of native Kubernetes manifests (known as YAML) to be defined as a Template.
