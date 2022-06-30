# cosmo-dashboard UI 

## How to create this project

```
git clone https://github.com/cosmo-workspace/cosmo
cd cosmo

WEBUI=$(pwd)/web/dashboard-ui

npx create-react-app --template=typescript web/dashboard-ui

cd $WEBUI
               
yarn add \
  @mui/material @emotion/react @emotion/styled \
  @mui/icons-material \
  @types/react-router-dom react-router-dom \
  react-hook-form \
  axios \
  notistack \
  copy-to-clipboard \
  react-error-boundary \
  http-proxy-middleware \
  @testing-library/react-hooks \
  @emotion/jest

  default
  "dependencies": {
    "@testing-library/jest-dom": "^5.11.4",
    "@testing-library/react": "^11.1.0",
    "@testing-library/user-event": "^12.1.10",
    "@types/jest": "^26.0.15",
    "@types/node": "^12.0.0",
    "@types/react": "^17.0.0",
    "@types/react-dom": "^17.0.0",
    "react": "^17.0.2",
    "react-dom": "^17.0.2",
    "react-scripts": "4.0.3",
    "typescript": "^4.1.2",
    "web-vitals": "^1.0.1"
  },

mkdir -p \
  $WEBUI/src/views/atoms \
  $WEBUI/src/views/molecules \
  $WEBUI/src/views/organisms \
  $WEBUI/src/views/templates \
  $WEBUI/src/views/pages \
  $WEBUI/src/hooks \
  $WEBUI/src/components \
  $WEBUI/src/rest


```

## How to start

```
$ cd web/dashboard-ui
$ yarn install && yarn start
```
