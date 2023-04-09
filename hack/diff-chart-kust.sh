#!/bin/bash

app=$1
no_remove_tmp=$2
shift; shift
HELM_OPT=$@
[[ $app != "controller-manager" && $app != "dashboard" ]] && echo "invalid args $1" && exit 1

readonly TMP_DIR=$(mktemp -d -p .)
echo "create tmp directory: $TMP_DIR"

if [[ $(which yq) == "" ]]; then
    curl --fail -L https://github.com/mikefarah/yq/releases/download/v4.9.6/yq_linux_amd64 -o $TMP_DIR/yq || exit 1
    YQ=$TMP_DIR/yq
    chmod +x $YQ
else
    YQ=yq
fi

kustomize build config/$app/ > $TMP_DIR/kust.yaml

helm template cosmo -n cosmo-system $HELM_OPT ./charts/cosmo-$app/ > $TMP_DIR/helm.yaml

kDocIndex=$($YQ eval 'documentIndex' $TMP_DIR/kust.yaml | tail -1)
hDocIndex=$($YQ eval 'documentIndex' $TMP_DIR/helm.yaml | tail -1)

for ((ki=0; ki<$kDocIndex+1; ki++)) {
    kKind=$($YQ e "select(di == $ki) | .kind" $TMP_DIR/kust.yaml)
    kName=$($YQ e "select(di == $ki) | .metadata.name" $TMP_DIR/kust.yaml)
    found=0
    for ((hi=0; hi<$hDocIndex+1; hi++)) {
        hKind=$($YQ e "select(di == $hi) | .kind" $TMP_DIR/helm.yaml)
        hName=$($YQ e "select(di == $hi) | .metadata.name" $TMP_DIR/helm.yaml)

        # echo "$kKind == $hKind && $kName == $hName"
        if [[ $kKind == $hKind ]] && [[ $kName == $hName ]]; then
            found=1
            $YQ e "select(di == $ki)" $TMP_DIR/kust.yaml > $TMP_DIR/kust.yaml.$kKind.$kName
            $YQ e "select(di == $hi)" $TMP_DIR/helm.yaml > $TMP_DIR/helm.yaml.$hKind.$hName

            kHeader=$(printf "%-45s" "$ki $kKind $kName" | cut -c 1-45)
            hHeader=$(printf "%-50s" "$hi $hKind $hName" | cut -c 1-50)
            printf "+- SHOW DIFF | %-45s | %-50s |\n" "$kHeader" "$hHeader"

            echo "+- SHOW DIFF |---------------- KUSTOMIZE --------------------|---------------- HELM CHART ------------------------|"
            sdiff $TMP_DIR/kust.yaml.$kKind.$kName $TMP_DIR/helm.yaml.$hKind.$hName
            echo "+- DIFF END  |---------------- KUSTOMIZE --------------------|---------------- HELM CHART ------------------------|"
        fi
    }
    [[ $found == 0 ]] && echo "+- $ki $kKind $kName | not found "
    echo
}


[[ $no_remove_tmp == "" ]] && echo "remove tmp directory: $TMP_DIR" && rm -rf $TMP_DIR