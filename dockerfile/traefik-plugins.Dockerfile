### see https://traefik.io/blog/using-private-plugins-in-traefik-proxy-2-5/
FROM golang:1.20

WORKDIR /cosmo

COPY . .

RUN make -C traefik/plugins/cosmo-workspace/cosmoauth vendor

FROM alpine:3
COPY --from=0 /cosmo/traefik/plugins/ /plugins-local/src/github.com/
