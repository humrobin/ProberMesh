package agent

import (
	"context"
	"github.com/sirupsen/logrus"
	"probermesh/pkg/pb"
	"time"
)

const (
	healthCheckInterval = time.Duration(10) * time.Second
)

type healthCheck struct {
	r          *rpcCli
	selfRegion string
	selfAddr   string
	ready      chan struct{}

	cancel context.Context
}

func newHealthCheck(ctx context.Context, r *rpcCli, ready chan struct{}) *healthCheck {
	return &healthCheck{
		r:          r,
		selfRegion: tm.selfRegion,
		selfAddr:   getLocalIP(),
		cancel:     ctx,
		ready:      ready,
	}
}

func (h *healthCheck) report() {
	ticker := time.NewTicker(healthCheckInterval)
	defer ticker.Stop()

	do := func() {
		var msg string
		err := h.r.Call(
			"Server.Report",
			pb.ReportReq{
				IP:     h.selfAddr,
				Region: h.selfRegion,
			},
			&msg,
		)
		if err != nil {
			logrus.Errorln("rpc report failed ", err)
		}
	}

	do()
	close(h.ready)
	for {
		select {
		case <-h.cancel.Done():
			return
		case <-ticker.C:
			do()
		}
	}
}
