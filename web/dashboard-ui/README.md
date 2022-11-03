# cosmo-dashboard UI 

## How to create this project

```
git clone https://github.com/cosmo-workspace/cosmo
cd cosmo

WEBUI=$(pwd)/web/dashboard-ui

npm create vite@latest -- dashboard-ui --template react-ts

cd $WEBUI

yarn add -D \
  @mui/material @mui/icons-material \
  @emotion/react @emotion/styled \
  @types/react-router-dom react-router-dom \
  axios \
  notistack \
  react-error-boundary \
  react-hook-form \
  copy-to-clipboard \
  web-vitals \
  @testing-library/react \
  @testing-library/user-event \
  @testing-library/jest-dom \
  @testing-library/react-hooks \
  @emotion/jest \
  jest @types/jest ts-jest jest-environment-jsdom

  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0"
  },
  "devDependencies": {
    "@types/react": "^18.0.22",
    "@types/react-dom": "^18.0.7",
    "@vitejs/plugin-react": "^2.2.0",
    "typescript": "^4.6.4",
    "vite": "^3.2.0"
  }

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
