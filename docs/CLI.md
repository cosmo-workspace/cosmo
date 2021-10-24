# cosmoctl

A command line utirities to manage COSMO resources.
It is able to do the same thing +α with COSMO Dashboard by CLI.

## Difference with COSMO Dashboard 
COSMO Dashboard has the own authentication but cosmoctl use Kubernetes RBAC for authentication.

It is because all of the COSMO resources are Kubernetes resources actually.
Even if one who uses COSMO User is not `cosmo-admin` in COSMO auhtentication,  COSMO User auhtentication is meanless when he/she is granted as `cluster-admin` in Kubernetes RBAC and can access all Kubernetes resources by Kubectl.

This means `cosmoctl` can create COSMO User (Kubernetes Namespace, acctually) even if there is NO User in COSMO Dashboard.

## Command Details

```
$ cosmoctl
Command line tool to manipulate comso
Complete documentation is available at http://github.com/cosmo-workspace/cosmo

MIT 2021 cosmo-workspace/cosmo

Usage:
  cosmoctl [command]

Available Commands:
  help        Help about any command
  template    Manipulate Template
  user        Manipulate UserNamespace
  version     Print the version number
  workspace   Manipulate Workspace

Flags:
      --context string      kube-context (default: current context)
  -h, --help                help for cosmoctl
      --kubeconfig string   kubeconfig file path (default: $HOME/.kube/config)
      --v int               log level (default: 0)

Use "cosmoctl [command] --help" for more information about a command.
```

>TODO