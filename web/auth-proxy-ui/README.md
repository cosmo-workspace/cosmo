# cosmo-auth-proxy login UI

## How to create this project

```
git clone https://github.com/cosmo-workspace/cosmo
cd cosmo/

WEBUI=$(pwd)/web/auth-proxy-ui

cd web/

npm create vite@latest -- auth-proxy-ui --template react-ts

cd $WEBUI

yarn add \
  @mui/material @emotion/react @emotion/styled \
  @mui/icons-material \
  react-error-boundary

# yarn add -D \
#  @bufbuild/protoc-gen-connect-web \
#  @bufbuild/protoc-gen-es

```

## How to start

```
cd web/auth-proxy-ui
yarn install && yarn dev
```

## How to Proxy test

```
# build
cd web/auth-proxy-ui
yarn build --base=/proxy-driver-test --outDir=build_test

../../hack/download-certs.sh dashboard-server-cert cosmo-system

go run ../../hack/echo-server/main.go &
go run ../../hack/proxy-driver/main.go --port=9999 --target-port=8888 --user=[COSMO_USER_ID] --auth-ui=./build_test/ --auth-url=http://cosmo-dashboard.cosmo-system.svc.cluster.local:8443