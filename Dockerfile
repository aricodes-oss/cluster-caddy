FROM caddy:builder as build

# Local replacement for darkweak/storages/redis/caddy: upstream opens a new
# rueidis client on every config reload and never closes the old one.
# See patches/storages-redis-caddy/redis.go.
COPY patches /src/patches

RUN xcaddy build \
  --with github.com/caddy-dns/cloudflare \
  --with github.com/mholt/caddy-dynamicdns \
  --with github.com/lucaslorentz/caddy-docker-proxy/v2 \
  --with github.com/caddyserver/cache-handler \
  --with github.com/mholt/caddy-ratelimit \
  --with pkg.jsn.cam/caddy-defender \
  --with github.com/darkweak/storages/redis/caddy=/src/patches/storages-redis-caddy \
  --with github.com/pberkel/caddy-storage-redis

FROM caddy:latest

COPY --from=build /usr/bin/caddy /usr/bin/caddy
CMD ["caddy", "docker-proxy"]
