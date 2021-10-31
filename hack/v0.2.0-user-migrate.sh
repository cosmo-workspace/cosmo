#!/bin/bash

for i in $(kubectl get ns -o jsonpath='{.items[*].metadata.name}'); do
    case $i in
        cosmo-user-*)
            ;;
        *)
            continue
            ;;
    esac

    echo "* checking $i is non managed cosmo user-namespace..."
    
    owner=$(kubectl get ns $i -o jsonpath='{.metadata.ownerReferences[0].kind}')
    if [[ ! -z $owner ]]; then
        echo "managed. skip"
        continue
    fi

    echo "not managed. creat new user"

    # kubectl get ns $i -o yaml
    user_id=$(kubectl get ns $i -o jsonpath='{.metadata.labels.cosmo\/user-id}')
    display_name=$(kubectl get ns $i -o jsonpath='{.metadata.annotations.cosmo\/user-name}')
    user_role=$(kubectl get ns $i -o jsonpath='{.metadata.annotations.cosmo\/user-role}')
    auth_type=$(kubectl get ns $i -o jsonpath='{.metadata.annotations.cosmo\/auth-type}')

    echo "creating new user...id=$user_id, display_name=$display_name, user_role=$user_role, auth_type=$auth_type"

    cat <<EOF | kubectl create -f - -o yaml
apiVersion: workspace.cosmo-workspace.github.io/v1alpha1
kind: User
metadata:
  name: $user_id
spec:
  displayName: "$display_name"
  role: "$user_role"
  authType: "$auth_type"
EOF

done