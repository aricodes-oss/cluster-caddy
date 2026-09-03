FROM caddy:builder as build

RUN xcaddy build \
  --with github.com/caddy-dns/cloudflare \
  --with github.com/mholt/caddy-dynamicdns \
  --with github.com/lucaslorentz/caddy-docker-proxy/v2 \
  --with github.com/caddyserver/cache-handler \
  --with github.com/mholt/caddy-ratelimit \
  --with pkg.jsn.cam/caddy-defender \
  --with github.com/darkweak/storages/redis/caddy \
  --with github.com/pberkel/caddy-storage-redis

FROM caddy:latest

COPY --from=build /usr/bin/caddy /usr/bin/caddy
CMD ["caddy", "docker-proxy"]
