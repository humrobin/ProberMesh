package agent

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/toolkits/pkg/net/gobrpc"
	"probermesh/config"
	"probermesh/pkg/pb"
	"time"
)

type targetManager struct {
	r               *gobrpc.RPCClient
	targets         map[string][]*config.ProberConfig
	refreshInterval time.Duration
	currents        map[string]*proberJob

	ready chan struct{}
}

var tm *targetManager

func GetTM() *targetManager {
	return tm
}

func initTargetManager(interval int, r *gobrpc.RPCClient) *targetManager {
	tm = &targetManager{
		targets:         make(map[string][]*config.ProberConfig),
		refreshInterval: time.Duration(interval) * time.Second,
		r:               r,
		ready:           make(chan struct{}),
		currents:        make(map[string]*proberJob),
	}
	return tm
}

func (t *targetManager) start(ctx context.Context) {
	// 定时获取targets
	go t.refresh(ctx)

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
				sourceRegion: "cn-shanghai",
				targetRegion: region,
				r:            t.r,
			}
			pj.run()
			t.currents["cn-shanghai"+region+tt.ProberType] = pj
		}
	}
}

func (t *targetManager) refresh(ctx context.Context) {
	t.getTargets()
	close(t.ready)

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(t.refreshInterval):
			t.getTargets()
		}
	}
}

func (t *targetManager) getTargets() {
	resp := new(pb.TargetPoolResp)
	err := t.r.Call(
		"Server.GetTargetPool",
		pb.TargetPoolReq{SourceRegion: "cn-shanghai"},
		resp,
	)
	if err != nil {
		logrus.Errorln("get targets failed ", err)
		return
	}

	t.targets = resp.Targets

	//for region, tg := range t.targets {
	//	for _, tt := range tg {
	//		fmt.Println(region, tt.ProberType, tt.Targets)
	//	}
	//}
}
