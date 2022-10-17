package agent

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"probermesh/config"
	"probermesh/pkg/pb"
	"time"
)

type targetManager struct {
	r               *rpcCli
	targets         map[string][]*config.ProberConfig
	refreshInterval time.Duration
	syncInterval    time.Duration
	currents        map[string]struct{}
	selfRegion      string

	ready, beforeReady chan struct{}
}

var tm *targetManager

func NewTargetManager(
	region string,
	pInterval,
	sInterval time.Duration,
	r *rpcCli,
	br chan struct{},
) *targetManager {
	tm = &targetManager{
		targets:         make(map[string][]*config.ProberConfig),
		refreshInterval: pInterval,
		syncInterval:    sInterval,
		r:               r,
		ready:           make(chan struct{}),
		beforeReady:     br,
	}

	// 如果指定了region,使用指定
	// 未指定region,自动获取
	if len(region) > 0 {
		tm.selfRegion = region
	} else {
		tm.selfRegion = getSelfRegion()
	}
	return tm
}

func (t *targetManager) start(ctx context.Context) {
	<-t.beforeReady

	// 定时获取targets
	go t.sync(ctx)

	<-t.ready

	// 定时探测
	t.prober()
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(t.refreshInterval):
			t.prober()
		}
	}
}

func (t *targetManager) prober() {
	for region, tg := range t.targets {
		for _, tt := range tg {
			pj := &proberJob{
				proberType:   tt.ProberType,
				targets:      tt.Targets,
				sourceRegion: t.selfRegion,
				targetRegion: region,
				r:            t.r,
			}
			pj.run()
		}
	}
}

func (t *targetManager) sync(ctx context.Context) {
	t.getTargets()
	close(t.ready)

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(t.syncInterval):
			t.getTargets()
		}
	}
}

func (t *targetManager) getTargets() {
	resp := new(pb.TargetPoolResp)
	err := t.r.Call(
		"Server.GetTargetPool",
		pb.TargetPoolReq{SourceRegion: t.selfRegion},
		resp,
	)
	if err != nil {
		logrus.Errorln("get targets failed ", err)
		return
	}

	var msg string
	for region, pcs := range resp.Targets {
		msg += fmt.Sprintf("[region == %s]|[targetLens == %d] ", region, len(pcs))
	}
	logrus.Warnln("agent get current target list msg: ", msg)
	t.targets = resp.Targets
}
