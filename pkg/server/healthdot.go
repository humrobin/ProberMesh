package server

import (
	"context"
	"strings"
	"sync"
	"time"
)

type healthDot struct {
	expires      time.Duration
	agentPool    map[string]time.Time
	discoverPool map[string]map[string]struct{}

	cancel context.Context
	ready  chan struct{}
	m      sync.Mutex
}

var hd *healthDot

func newHealthDot(ctx context.Context, expires time.Duration, ready chan struct{}) *healthDot {
	hd = &healthDot{
		expires:      expires,
		agentPool:    make(map[string]time.Time),
		discoverPool: make(map[string]map[string]struct{}),
		cancel:       ctx,
		ready:        ready,
	}
	return hd
}

func (h *healthDot) report(region, ip string) {
	h.m.Lock()
	defer h.m.Unlock()
	h.agentPool[region+defaultKeySeparator+ip] = time.Now()

	// 将上报的agent region和ip存入
	if ipm, ok := h.discoverPool[region]; ok {
		ipm[ip] = struct{}{}
		//ips = append(ips, ip)
	} else {
		h.discoverPool[region] = map[string]struct{}{ip: {}}
	}

	if h.ready != nil {
		close(h.ready)
		h.ready = nil
	}
}

func (h *healthDot) dot() {
	ticker := time.NewTicker(h.expires)
	defer ticker.Stop()

	for {
		select {
		case <-h.cancel.Done():
			return
		case <-ticker.C:
			now := time.Now()
			for agent, tm := range h.agentPool {
				meta := strings.Split(agent, defaultKeySeparator)
				region, ip := meta[0], meta[1]
				if now.Sub(tm) > h.expires {
					agentHealthCheckGaugeVec.WithLabelValues(region, ip).Set(0)

					h.m.Lock()
					delete(h.agentPool, agent)
					delete(h.discoverPool[region], ip)
					h.m.Unlock()
				} else {
					agentHealthCheckGaugeVec.WithLabelValues(region, ip).Set(1)
				}
			}
		}
	}
}

func getDiscoverPool() map[string]map[string]struct{} {
	return hd.discoverPool
}
