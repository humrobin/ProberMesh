package agent

import (
	"context"
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
	currents        map[string]*proberJob
	selfRegion      string

	ready chan struct{}
}

var tm *targetManager

func NewTargetManager(pInterval, sInterval time.Duration, r *rpcCli) *targetManager {
	tm = &targetManager{
		targets:         make(map[string][]*config.ProberConfig),
		refreshInterval: pInterval,
		syncInterval:    sInterval,
		r:               r,
		ready:           make(chan struct{}),
		currents:        make(map[string]*proberJob),
		selfRegion:      getSelfRegion(),
	}
	return tm
}

func (t *targetManager) start(ctx context.Context) {
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
			t.currents["cn-shanghai"+region+tt.ProberType] = pj
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

	t.targets = resp.Targets
}
