[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/cosmo)](https://artifacthub.io/packages/search?repo=cosmo)

# COSMO
helm chart for [COSMO](https://github.com/cosmo-workspace/cosmo)

For general installation instructions, see [GETTING-STARTED.md](https://github.com/cosmo-workspace/cosmo/blob/main/docs/GETTING-STARTED.md).

## Install options

Example

```sh
helm upgrade --install -n cosmo-system --create-namespace cosmo cosmo/cosmo \
  --set domain=example.com
```

See detail in [`values.yaml`](https://github.com/cosmo-workspace/cosmo/blob/main/charts/cosmo/values.yaml)
