### see https://traefik.io/blog/using-private-plugins-in-traefik-proxy-2-5/
FROM alpine:3

RUN apk add --update git
RUN PLUGIN=github.com/wiltonsr/ldapAuth && git clone https://${PLUGIN}.git /plugins-local/src/${PLUGIN} --depth 1
# COPY traefik/plugins/ plugins-local/src/github.com/

FROM alpine:3
COPY --from=0 /plugins-local /plugins-local
