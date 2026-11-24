// Patched copy of github.com/darkweak/storages/redis/caddy v0.0.22.
//
// Upstream's Provision calls redis.Factory -- a brand-new rueidis client --
// every time a `cache` handler is provisioned, which is once per handler per
// config load. Nothing ever closes the previous one: SouinApp.Stop only clears
// the registry map, so each reload strands a client that rueidis keeps
// PINGing forever. In our swarm the controller pushes on every task move, and
// one bad day leaked ~6,000 connections to valkey.
//
// Closing the old client on Cleanup is NOT safe: Souin's mapping-eviction
// goroutine (deduped by storer UUID in a process-global registry) and the
// surrogate storage both keep a reference to the first storer ever built.
// Instead, keep one storer per distinct configuration for the life of the
// process and hand the same one back on every reload. rueidis reconnects on
// its own, so a long-lived client is the intended usage.
package caddy

import (
	"encoding/json"
	"net/http"
	"sync"

	caddy "github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/darkweak/storages/core"
	"github.com/darkweak/storages/redis"
)

const moduleName = "redis"

var (
	storersMu sync.Mutex
	storers   = map[string]core.Storer{}
)

// Redis storage.
type Redis struct {
	// Keep the handler configuration.
	core.Configuration
}

//nolint:gochecknoinits
func init() {
	caddy.RegisterModule(Redis{})
}

// CaddyModule returns the Caddy module information.
func (Redis) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "storages.cache.redis",
		New: func() caddy.Module { return new(Redis) },
	}
}

// Provision to do the provisioning part.
func (b *Redis) Provision(ctx caddy.Context) error {
	logger := ctx.Logger(b)

	k, err := json.Marshal(b.Configuration)
	if err != nil {
		return err
	}

	storersMu.Lock()
	defer storersMu.Unlock()

	storer, ok := storers[string(k)]
	if !ok {
		storer, err = redis.Factory(b.Configuration.Provider, logger.Sugar(), b.Configuration.Stale)
		if err != nil {
			return err
		}

		storers[string(k)] = storer
	}

	// Re-register every time: SouinApp.Stop wipes the registry on each reload.
	core.RegisterStorage(storer)

	return nil
}

func (b *Redis) ServeHTTP(rw http.ResponseWriter, rq *http.Request, next caddyhttp.Handler) error {
	return next.ServeHTTP(rw, rq)
}

// Interface guards.
var (
	_ caddy.Provisioner           = (*Redis)(nil)
	_ caddyhttp.MiddlewareHandler = (*Redis)(nil)
)
