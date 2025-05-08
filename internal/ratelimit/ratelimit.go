package ratelimit

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// Реализация BucketToken

// ClientRateProvider возвращает параметры для BucketToken в зависимости от IP клиента
type ClientRateProvider interface {
	GetLimitForClient(clientIp string) (uint64, uint64)
}

type MockLimitProvider struct {
	Cap        uint64
	RefillRate uint64
}

func (m *MockLimitProvider) GetLimitForClient(clientIp string) (uint64, uint64) {
	return m.Cap, m.RefillRate
}

// Bucket это структура, которая хранит токены клиента и говорит, есть ли свободные токены у клиента
type Bucket struct {
	capacity   uint64
	tokens     uint64
	refillRate uint64
	lastRefill int64
	mu         sync.Mutex
}

func NewBucket(capacity, refillRate uint64) *Bucket {
	now := time.Now().UnixNano()
	return &Bucket{capacity: capacity, tokens: capacity, refillRate: refillRate, lastRefill: now}
}

func (b *Bucket) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now().UnixNano()
	delta := now - b.lastRefill
	add := uint64(delta) * b.refillRate / uint64(time.Second)
	if add > 0 {
		b.tokens = min(b.capacity, b.tokens+add)
		b.lastRefill = now
	}
	if b.tokens == 0 {
		return false
	}
	b.tokens--
	return true
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

// rateLimiter определяет, пройдёт ли дальше запрос в зависимости от состояния Bucket клиента
type rateLimiter struct {
	buckets  sync.Map
	provider ClientRateProvider
}

func NewMiddleware(provider ClientRateProvider) func(http.Handler) http.Handler {
	rl := &rateLimiter{provider: provider}
	return rl.limit
}

func (rl *rateLimiter) limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		client := clientIP(r)
		capacity, rate := rl.provider.GetLimitForClient(client)
		bIface, _ := rl.buckets.LoadOrStore(client, NewBucket(capacity, rate))
		bucket := bIface.(*Bucket)
		if !bucket.Allow() {
			w.Header().Set("Retry-After", "1")
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
