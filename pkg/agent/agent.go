package agent

import (
	"context"
	"github.com/oklog/run"
	"github.com/sirupsen/logrus"
	"log"
	"probermesh/pkg/util"
)

type ProberMeshAgentOption struct {
	Addr, PInterval, SInterval string
}

func BuildAgentMode(ao *ProberMeshAgentOption) {
	if len(ao.Addr) == 0 {
		log.Fatal("server addr must be set")
	}

	pDuration, err := util.ParseDuration(ao.PInterval)
	if err != nil {
		logrus.Errorln("parse prober duration flag failed ", err)
		return
	}

	sDuration, err := util.ParseDuration(ao.SInterval)
	if err != nil {
		logrus.Errorln("parse sync duration flag failed ", err)
		return
	}

	ctxAll, cancelAll := context.WithCancel(context.Background())

	cli := initRpcCli(ctxAll, ao.Addr)

	if err != nil {
		logrus.Errorln("agent get cli failed ", err)
		return
	}

	var g run.Group
	{
		// 定时拉取mesh poll
		manager := NewTargetManager(pDuration, sDuration, cli)
		g.Add(func() error {
			manager.start(ctxAll)
			return nil
		}, func(err error) {
			cancelAll()
		})
	}

	g.Run()
}
