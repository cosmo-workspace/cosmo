# cosmo-auth-proxy login UI

## How to create this project

```
git clone https://github.com/cosmo-workspace/cosmo
cd cosmo

WEBUI=$(pwd)/web/auth-proxy-ui

npx create-react-app --template=typescript web/auth-proxy-ui

cd $WEBUI

yarn add \
  @mui/material @emotion/react @emotion/styled \
  @mui/icons-material \
  react-error-boundary

mkdir -p \
  $WEBUI/src/views/atoms \
  $WEBUI/src/views/pages \
  $WEBUI/src/components \
```

## How to start

```
cd web/auth-proxy-ui
yarn install && yarn start
```

## How to Proxy test

```
# build
cd web/auth-proxy-ui
PUBLIC_URL=/proxy-driver-test BUILD_PATH=build_test yarn build

../../hack/download-certs.sh dashboard-server-cert cosmo-system

go run ../../hack/proxy-driver/main.go --port=9999 --target-port=[BACKEND_PORT] --user=[COSMO_USER_ID] --auth-ui=./build_test/ --auth-url=[DASHBOARD_URL]
