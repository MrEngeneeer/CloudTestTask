package server

import (
	"net/http"
	"slices"
	"time"

	"balancer/internal/balancer"
	"balancer/internal/config"
	"balancer/internal/health"
	"balancer/internal/logging"
	"balancer/internal/proxy"
	"balancer/internal/ratelimit"
)

// Сборка сервера

func Run(cfg *config.Config) error {
	hc := health.New(cfg.Backends, time.Duration(cfg.HealthCheck.IntervalSec)*time.Second, cfg.HealthCheck.Path)
	hc.Start()

	rr, _ := balancer.NewRoundRobin(cfg.Backends)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		live := hc.Alive()
		if len(live) == 0 {
			http.Error(w, "no backends available", http.StatusServiceUnavailable)
			return
		}
		targetURL, _ := rr.NextBackend()
		for !slices.Contains(live, targetURL.String()) {
			logging.Std.Info("Backend %s is non available, skip", targetURL.String())
			targetURL, _ = rr.NextBackend()
		}
		logging.Std.Info("Backend %s selected", targetURL.String())
		p := proxy.NewReverseProxy(targetURL)
		p.ServeHTTP(w, r)
	})

	handler := logging.Std.
		Middleware(
			ratelimit.NewMiddleware(&ratelimit.MockLimitProvider{Cap: cfg.RateLimit.Capacity, RefillRate: cfg.RateLimit.RefillRate})(
				mux,
			),
		)

	return http.ListenAndServe(cfg.ListenAddr, handler)
}
