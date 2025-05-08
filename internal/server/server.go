package server

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
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

	rateProvider := ratelimit.NewDefaultLimitProvider(cfg.RateLimit.Capacity, cfg.RateLimit.RefillRate)
	// endpoint для добавления особых лимитов клиента
	mux.HandleFunc("/clients", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var req struct {
				ClientIp string `json:"client_ip"`
				ratelimit.Limit
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			err := rateProvider.AddClient(req.ClientIp, ratelimit.Limit{
				Capacity: req.Capacity,
				Rate:     req.Rate,
			})
			if err != nil {
				http.Error(w, "error in client adding", http.StatusInternalServerError)
			}
			w.WriteHeader(http.StatusCreated)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	// endpoint для удаления особых лимитов клиента
	mux.HandleFunc("/clients/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) != 2 || parts[0] != "clients" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		clientIp := parts[1]
		err := rateProvider.DeleteClient(clientIp)
		if err != nil {
			http.Error(w, "error in client deleting", http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	handler := logging.Std.
		Middleware(
			ratelimit.NewMiddleware(rateProvider)(
				mux,
			),
		)

	return http.ListenAndServe(cfg.ListenAddr, handler)
}
