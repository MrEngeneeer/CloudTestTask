package balancer

import (
	"errors"
	"net/url"
	"sync/atomic"
)

// Реализация алгоритма балансировки

type Balancer interface {
	NextBackend() (*url.URL, error)
	Backends() []*url.URL
}

type RoundRobin struct {
	backends []*url.URL
	idx      uint64
}

func NewRoundRobin(rawBackends []string) (*RoundRobin, error) {
	if len(rawBackends) == 0 {
		return nil, errors.New("no backends provided")
	}
	urls := make([]*url.URL, len(rawBackends))
	for i, b := range rawBackends {
		u, err := url.Parse(b)
		if err != nil {
			return nil, err
		}
		urls[i] = u
	}
	return &RoundRobin{backends: urls}, nil
}

func (r *RoundRobin) NextBackend() (*url.URL, error) {
	n := atomic.AddUint64(&r.idx, 1)
	if len(r.backends) == 0 {
		return nil, errors.New("no backends available")
	}
	return r.backends[(int(n)-1)%len(r.backends)], nil
}

func (r *RoundRobin) Backends() []*url.URL {
	return r.backends
}
