all: user-addon

user-addon:
	cat job.yaml | cosmoctl template generate --name eks-iamserviceaccount \
		--user-addon \
		--required-vars SERVICE_ACCOUNT:default,CLUSTER_NAME \
		--desc 'create IAM Role for Service Account in user namespace' > user-addon-eks-iamserviceaccount.yaml