package v1

import (
	"os"
	"sync"
)

// Environment is where libtunnel reads its environment variables and writes
// the spec handoff. The default is the process environment. An application
// that embeds libtunnel and configures every tunnel in code can install
// Isolated instead, so a variable inherited from whoever started the process
// (LIBTUNNEL_LOCAL_URL, LIBTUNNEL_SPEC, ...) can neither redirect a tunnel nor
// receive a tunnel's secret.
//
// It covers the LIBTUNNEL_* variables and the token fallback. Proxy settings
// (HTTPS_PROXY and friends) stay with the process environment: they describe
// the network the process runs in, not a tunnel.
type Environment interface {
	// LookupEnv reports a variable's value and whether it is set, like
	// os.LookupEnv.
	LookupEnv(key string) (string, bool)
	// Setenv sets a variable, like os.Setenv. libtunnel calls it only to
	// export a minted spec for children (SpecEnv, HostnameEnv).
	Setenv(key, value string) error
}

type processEnv struct{}

func (processEnv) LookupEnv(key string) (string, bool) { return os.LookupEnv(key) }
func (processEnv) Setenv(key, value string) error      { return os.Setenv(key, value) }

type isolatedEnv struct{}

func (isolatedEnv) LookupEnv(string) (string, bool) { return "", false }
func (isolatedEnv) Setenv(string, string) error     { return nil }

// Isolated is an Environment with nothing in it: every variable reads as
// unset and every write is dropped. Code is the only configuration, and no
// spec is exported for children.
var Isolated Environment = isolatedEnv{}

var (
	envMu sync.RWMutex
	env   Environment = processEnv{}
)

// SetEnvironment installs the Environment every tunnel in this process reads
// from and exports to; nil restores the process environment. Call it before
// the first tunnel starts: each variable is read where its knob takes effect,
// so a tunnel already running may have read some under the old one.
func SetEnvironment(e Environment) {
	if e == nil {
		e = processEnv{}
	}
	envMu.Lock()
	env = e
	envMu.Unlock()
}

func current() Environment {
	envMu.RLock()
	defer envMu.RUnlock()
	return env
}

// LookupEnv reads a variable from the installed Environment.
func LookupEnv(key string) (string, bool) { return current().LookupEnv(key) }

// Getenv reads a variable from the installed Environment, empty when unset.
func Getenv(key string) string {
	v, _ := current().LookupEnv(key)
	return v
}

// Setenv writes a variable to the installed Environment.
func Setenv(key, value string) error { return current().Setenv(key, value) }
