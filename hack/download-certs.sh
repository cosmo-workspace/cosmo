#/bin/bash

secert=$1
[[ $secert == "" ]] && echo "invalid args: ./download-certs.sh SECRET_NAME NAMESPACE" && exit 9
namespace=$2
[[ $namespace == "" ]] && namespace="cosmo-system"

kubectl get secret -n $namespace $secert -o jsonpath='{.data.tls\.crt}' | base64 -d > tls.crt
kubectl get secret -n $namespace $secert -o jsonpath='{.data.tls\.key}' | base64 -d > tls.key
kubectl get secret -n $namespace $secert -o jsonpath='{.data.ca\.crt}' | base64 -d > ca.crt
echo DONE
