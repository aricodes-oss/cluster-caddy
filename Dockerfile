FROM caddy:builder as build

RUN xcaddy build --with github.com/caddy-dns/cloudflare --with github.com/mholt/caddy-dynamicdns --with github.com/lucaslorentz/caddy-docker-proxy/v2

FROM caddy:latest

COPY --from=build /usr/bin/caddy /usr/bin/caddy
CMD ["caddy", "docker-proxy"]
