# eks-irsa-useraddon example

Example COSMO UserAddon Template for creating serviceaccount associated with IAM Role

```sh
# fetch Cluster's OIDC Provider info
OIDC_PROVIDER=$(aws eks describe-cluster --name $CLUSTER --query "cluster.identity.oidc.issuer" --output text | sed -e "s/^https:\/\///")

# replace OIDC Provider info in cloudformation template
sed -e "s/OIDC_PROVIDER/$OIDC_PROVIDER/" iamrole.cloudformation.yaml > iamrole.cloudformation.yaml.tmp

# create stack
aws cloudformation create-stack --stack-name cosmo-test-user-addon --template-body file://iamrole.cloudformation.yaml.tmp --capabilities CAPABILITY_NAMED_IAM

# create COSMO Template
kubectl create -f eks-irsa-useraddon.yaml

# create COSMO User with "eks-irsa-useraddon" UserAddon
cosmoctl user create sample-user --addon eks-irsa-useraddon
```
