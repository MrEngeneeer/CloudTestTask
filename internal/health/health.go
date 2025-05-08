package health

import (
	"net/http"
	"sync"
	"time"
)

// Алгоритм для проверки состояния доступных бэкендов

// Checker с задаваемым интервалом проверяет состояние бэкендов по задаваемому пути и предоставляет "здоровые" бэкенды
type Checker struct {
	backends  []string
	status    map[string]bool
	checkPath string
	mu        sync.RWMutex
	interval  time.Duration
	client    *http.Client
}

func New(backends []string, interval time.Duration, checkPath string) *Checker {
	status := make(map[string]bool, len(backends))
	for _, b := range backends {
		status[b] = false
	}
	return &Checker{
		backends:  backends,
		status:    status,
		checkPath: checkPath,
		interval:  interval,
		client:    &http.Client{Timeout: 2 * time.Second},
	}
}

func (c *Checker) Start() {
	go func() {
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()
		for {
			c.checkAll()
			<-ticker.C
		}
	}()
}

func (c *Checker) checkAll() {
	for _, b := range c.backends {
		go c.check(b)
	}
}

func (c *Checker) check(backend string) {
	resp, err := c.client.Get(backend + c.checkPath)

	alive := err == nil && resp.StatusCode == http.StatusOK
	c.mu.Lock()
	c.status[backend] = alive
	c.mu.Unlock()
	if resp != nil {
		resp.Body.Close()
	}
}

func (c *Checker) Alive() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var alives []string
	for b, ok := range c.status {
		if ok {
			alives = append(alives, b)
		}
	}
	return alives
}
