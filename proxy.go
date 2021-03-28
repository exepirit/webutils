package webutils

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// Proxy is an HTTP handler that forward incoming client request
// and sends it to defined server.
//
// Proxy replaces the prefix defined in PathPrefix by TargetAddr path.
type Proxy struct {
	TargetAddr *url.URL
	PathPrefix string
}

// NewProxy returns new Proxy that routes requests with prefix
// to target server.
func NewProxy(target, prefix string) *Proxy {
	u, err := url.Parse(target)
	if err != nil {
		panic(err)
	}

	return &Proxy{
		TargetAddr: u,
		PathPrefix: prefix,
	}
}

func (proxy *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = strings.Replace(r.URL.Path, proxy.PathPrefix, proxy.TargetAddr.Path, 1)
	p := httputil.NewSingleHostReverseProxy(proxy.TargetAddr)
	p.ServeHTTP(w, r)
}
