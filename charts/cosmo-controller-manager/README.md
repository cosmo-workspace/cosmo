[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/cosmo)](https://artifacthub.io/packages/search?repo=cosmo)

# COSMO Controller Manager
helm chart for [COSMO Controller Manager](https://github.com/cosmo-workspace/cosmo)

For general installation instructions, see [GETTING-STARTED.md](https://github.com/cosmo-workspace/cosmo/blob/main/docs/GETTING-STARTED.md).

## Install options

Example

```sh
helm upgrade --install -n cosmo-system --create-namespace cosmo-controller-manager cosmo/cosmo-controller-manager --set logLevel=debug
```

| Option | Avairable values (default) | Description |
|:-------|:----------------|:------------|
| logLevel | ["info", "debug", 2(DEBUG_ALL) ] (info) | Loglevel for zap logger |
| enableCertManager | [true, false] (true) | Use cert-manager to gen cert for Admission Webhook. Or use helm inner function |

See detail in [`values.yaml`](https://github.com/cosmo-workspace/cosmo/blob/main/charts/cosmo-controller-manager/values.yaml)
