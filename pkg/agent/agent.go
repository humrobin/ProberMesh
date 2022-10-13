package agent

import (
	"context"
	"github.com/oklog/run"
	"github.com/sirupsen/logrus"
	"log"
	"probermesh/pkg/util"
)

type ProberMeshAgentOption struct {
	Addr, PInterval, SInterval, Region string
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
	var g run.Group
	{
		// 定时拉取mesh poll
		manager := NewTargetManager(
			ao.Region,
			pDuration,
			sDuration,
			cli,
		)
		g.Add(func() error {
			manager.start(ctxAll)
			return nil
		}, func(err error) {
			cancelAll()
		})
	}

	{
		// healthCheck
		g.Add(func() error {
			newHealthCheck(ctxAll, cli).report()
			return nil
		}, func(e error) {
			cancelAll()
		})
	}

	g.Run()
}
