# ----- traefik-plugin-builder ------
# https://traefik.io/blog/using-private-plugins-in-traefik-proxy-2-5/
FROM golang:1.22 as traefik-plugin-builder

WORKDIR /cosmo
COPY . .

RUN make -C traefik-plugins/src/github.com/cosmo-workspace/cosmoauth vendor

RUN cd traefik-plugins && tar zcvf traefik-plugins.tar.gz src/

# ----- builder ------
FROM golang:1.22 as builder

ENV GO111MODULE=on

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/ cmd/
COPY api/ api/
COPY pkg/ pkg/
COPY internal/ internal/

# Copy traefik-plugins embeded
COPY --from=traefik-plugin-builder /cosmo/traefik-plugins/traefik-plugins.tar.gz cmd/controller-manager/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager ./cmd/controller-manager/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
