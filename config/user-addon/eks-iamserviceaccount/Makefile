ACCOUNT_ID ?= $(shell aws sts get-caller-identity --query 'Account' --output text)

all: user-addon

user-addon:
	cat job.yaml | cosmoctl template generate --name eks-iamserviceaccount \
		--user-addon \
		--required-vars SERVICE_ACCOUNT:default,CLUSTER_NAME,AWS_REGION \
		--set-sysns-user-addon cosmo-system \
		--desc 'create IAM Role for Service Account in user namespace.' > user-addon-eks-iamserviceaccount.yaml

install: user-addon
ifndef EKS_CLUSTER_NAME
	$(error EKS_CLUSTER_NAME is not defined)
endif
	export AWS_PAGER="" aws iam create-policy --policy-name eksctl-policy --policy-document file://policy.json
	eksctl create iamserviceaccount \
		--cluster $(EKS_CLUSTER_NAME) \
		--name eksctl \
		--namespace cosmo-system \
		--attach-policy-arns arn:aws:iam::aws:policy/AWSCloudFormationFullAccess,arn:aws:iam::aws:policy/AmazonEC2FullAccess,arn:aws:iam::${ACCOUNT_ID}:policy/eksctl-policy
	kubectl create -f user-addon-eks-iamserviceaccount.yaml