# resource limitter example

Example COSMO UserAddon Template for limiting user resource consumptions by ResourceQuota

```sh
# create COSMO Template
kubectl create -f resource-limitter.yaml

# create COSMO User with "resource-limitter" UserAddon
cosmoctl user create sample-user --addon resource-limitter
```
