FROM golang:1.16 as base

ENV GO111MODULE=on

WORKDIR /cosmo

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

FROM base as builder

# Copy the go source
COPY cmd/ cmd/
COPY api/ api/
COPY pkg/ pkg/
COPY internal/ internal/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o auth-proxy ./cmd/auth-proxy/main.go

FROM node:14-alpine as ui-builder

# Create build environment
ENV PATH web/auth-proxy-ui/node_modules/.bin:$PATH
RUN mkdir -p web/auth-proxy-ui
WORKDIR /cosmo/web/auth-proxy-ui
# Copy the package.json and install
COPY web/auth-proxy-ui/package.json package.json
COPY web/auth-proxy-ui/tsconfig.json tsconfig.json
COPY web/auth-proxy-ui/yarn.lock yarn.lock
RUN yarn install

# Copy the source and build
COPY ./web/auth-proxy-ui .
RUN GENERATE_SOURCEMAP=false PUBLIC_URL=/cosmo-auth-proxy yarn build

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY --from=builder /cosmo/auth-proxy .
COPY --from=ui-builder /cosmo/web/auth-proxy-ui/build ./public

USER 65532:65532

ENTRYPOINT ["/app/auth-proxy"]
