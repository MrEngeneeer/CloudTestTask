package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// Прокси для перенаправления запроса (например на один из бэкендов)

func NewReverseProxy(target *url.URL) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)
	origDirector := proxy.Director
	proxy.Director = func(r *http.Request) {
		origDirector(r)
		r.Header.Set("X-Forwarded-Host", r.Host)
		r.Header.Set("X-Origin-Host", target.Host)
		r.Host = target.Host
	}
	return proxy
}
