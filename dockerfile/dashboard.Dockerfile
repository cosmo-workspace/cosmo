FROM golang:1.20 as base

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
COPY proto/ proto/
COPY internal/ internal/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dashboard ./cmd/dashboard/main.go
FROM node:16-alpine as ui-builder

# Create build environment
ENV PATH web/dashboard-ui/node_modules/.bin:$PATH
RUN mkdir -p web/dashboard-ui
WORKDIR /cosmo/web/dashboard-ui
# Copy the package.json and install
COPY web/dashboard-ui/package.json package.json
COPY web/dashboard-ui/tsconfig.json tsconfig.json
COPY web/dashboard-ui/yarn.lock yarn.lock
RUN yarn install

# Copy the source and build
COPY ./web/dashboard-ui .
RUN GENERATE_SOURCEMAP=false yarn build

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY --from=builder /cosmo/dashboard .
COPY --from=ui-builder /cosmo/web/dashboard-ui/dist ./public

USER 65532:65532

ENTRYPOINT ["/app/dashboard"]
