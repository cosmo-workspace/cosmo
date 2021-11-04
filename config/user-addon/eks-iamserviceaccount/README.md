# eks-iamserviceaccount

Before you use, create iamserviceaccount for the eksctl jobs.

```sh
ACCOUNT_ID=$(aws sts get-caller-identity --query 'Account' --output text)

aws iam create-policy --policy-name eksctl-policy --policy-document file://policy.json

eksctl create iamserviceaccount \
    --cluster YOUR_CLUSTER_NAME \
    --name eksctl \
    --namespace cosmo-system \
    --attach-policy-arns arn:aws:iam::aws:policy/AWSCloudFormationFullAccess,arn:aws:iam::aws:policy/AmazonEC2FullAccess,arn:aws:iam::${ACCOUNT_ID}:policy/eksctl-policy
```

Install

```
kubectl create -f user-addon-eks-iamserviceaccount.yaml
```