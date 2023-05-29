# amazon-efs-shared-filesystem example

Example COSMO UserAddon Template for creating shared Amazon EFS Filesystem PersistentVolume, StorageClass, PersistentVolumeClaim.

Amazon EFS CSI Driver is not support to create fixed path Amazon EFS Filesystem PersistentVolume dynamically for now.

COSMO UserAddon can create them by users.

```sh
# create COSMO Template
kubectl create -f efs-shared-filesystem.yaml.yaml

# create COSMO User with "efs-shared-filesystem" UserAddon
EFS_FILESYSTEM_ID=yours

cosmoctl user create sample-user --addon efs-shared-filesystem,EFS_FILESYSTEM_ID:$EFS_FILESYSTEM_ID
```
