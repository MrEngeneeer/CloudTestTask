// tests/integration/integration_test.go
package integration

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"balancer/internal/config"
	"balancer/internal/server"
)

func mockBackend(name string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from %s", name)
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "ok")
	})
	return httptest.NewServer(mux)
}

func sickMockBackend(name string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from %s", name)
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	return httptest.NewServer(mux)
}

func findFreePort() (string, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return "", err
	}
	defer l.Close()
	return l.Addr().String(), nil
}

func TestRunSmoke(t *testing.T) {
	b1 := mockBackend("b1")
	defer b1.Close()
	b2 := mockBackend("b2")
	defer b2.Close()

	portAddr, err := findFreePort()
	if err != nil {
		t.Fatal(err)
	}
	var cfg config.Config
	cfg.ListenAddr = portAddr
	cfg.Backends = []string{b1.URL, b2.URL}

	cfg.RateLimit.Capacity = 1000
	cfg.RateLimit.RefillRate = 1000

	cfg.HealthCheck.IntervalSec = 1
	cfg.HealthCheck.Path = "/health"

	go func() {
		if err := server.Run(&cfg); err != nil {
			fmt.Fprintln(os.Stderr, "server error:", err)
		}
	}()
	time.Sleep(1500 * time.Millisecond)

	resp, err := http.Get("http://" + portAddr + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
}

func TestSickBackend(t *testing.T) {
	b1 := sickMockBackend("b1")
	defer b1.Close()
	b2 := sickMockBackend("b2")
	defer b2.Close()

	portAddr, err := findFreePort()
	if err != nil {
		t.Fatal(err)
	}
	var cfg config.Config
	cfg.ListenAddr = portAddr
	cfg.Backends = []string{b1.URL, b2.URL}

	cfg.RateLimit.Capacity = 1000
	cfg.RateLimit.RefillRate = 1000

	cfg.HealthCheck.IntervalSec = 1
	cfg.HealthCheck.Path = "/health"

	go func() {
		if err := server.Run(&cfg); err != nil {
			fmt.Fprintln(os.Stderr, "server error:", err)
		}
	}()
	time.Sleep(1500 * time.Millisecond)

	resp, err := http.Get("http://" + portAddr + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
}

func TestOneSickBackend(t *testing.T) {
	b1 := sickMockBackend("b1")
	defer b1.Close()
	b2 := mockBackend("b2")
	defer b2.Close()

	portAddr, err := findFreePort()
	if err != nil {
		t.Fatal(err)
	}
	var cfg config.Config
	cfg.ListenAddr = portAddr
	cfg.Backends = []string{b1.URL, b2.URL}

	cfg.RateLimit.Capacity = 1000
	cfg.RateLimit.RefillRate = 1000

	cfg.HealthCheck.IntervalSec = 1
	cfg.HealthCheck.Path = "/health"

	go func() {
		if err := server.Run(&cfg); err != nil {
			fmt.Fprintln(os.Stderr, "server error:", err)
		}
	}()
	time.Sleep(1500 * time.Millisecond)

	resp, err := http.Get("http://" + portAddr + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
}

func TestRateLimit(t *testing.T) {
	b1 := mockBackend("b1")
	defer b1.Close()
	b2 := mockBackend("b2")
	defer b2.Close()

	portAddr, err := findFreePort()
	if err != nil {
		t.Fatal(err)
	}
	var cfg config.Config
	cfg.ListenAddr = portAddr
	cfg.Backends = []string{b1.URL, b2.URL}

	cfg.RateLimit.Capacity = 0
	cfg.RateLimit.RefillRate = 0

	cfg.HealthCheck.IntervalSec = 1
	cfg.HealthCheck.Path = "/health"

	go func() {
		if err := server.Run(&cfg); err != nil {
			fmt.Fprintln(os.Stderr, "server error:", err)
		}
	}()
	time.Sleep(1500 * time.Millisecond)

	resp, err := http.Get("http://" + portAddr + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
}

func BenchmarkBalancer(b *testing.B) {
	b1 := mockBackend("b1")
	defer b1.Close()
	b2 := mockBackend("b2")
	defer b2.Close()

	portAddr, err := findFreePort()
	if err != nil {
		b.Fatal(err)
	}
	var cfg config.Config
	cfg.ListenAddr = portAddr
	cfg.Backends = []string{b1.URL, b2.URL}

	cfg.RateLimit.Capacity = uint64(b.N)
	cfg.RateLimit.RefillRate = uint64(b.N)

	cfg.HealthCheck.IntervalSec = 1
	cfg.HealthCheck.Path = "/health"

	go func() {
		if err := server.Run(&cfg); err != nil {
			b.Fatalf("server error: %v", err)
		}
	}()
	time.Sleep(1500 * time.Millisecond)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := http.Get("http://" + portAddr + "/")
			if err != nil {
				b.Error(err)
				continue
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	})
}
