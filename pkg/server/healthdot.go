package server

import (
	"context"
	"strings"
	"sync"
	"time"
)

type healthDot struct {
	expires   time.Duration
	agentPool map[string]time.Time

	cancel context.Context
	m      sync.Mutex
}

var hd *healthDot

func newHealthDot(ctx context.Context, expires time.Duration) *healthDot {
	hd = &healthDot{
		expires:   expires,
		agentPool: make(map[string]time.Time),
		cancel:    ctx,
	}
	return hd
}

func (h *healthDot) report(region, ip string) {
	h.m.Lock()
	defer h.m.Unlock()
	h.agentPool[region+defaultKeySeparator+ip] = time.Now()
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
					h.m.Unlock()
				} else {
					agentHealthCheckGaugeVec.WithLabelValues(region, ip).Set(1)
				}
			}
		}
	}
}
