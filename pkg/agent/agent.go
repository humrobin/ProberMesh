package agent

import (
	"context"
	"github.com/oklog/run"
	"github.com/sirupsen/logrus"
	"log"
	"probermesh/pkg/util"
)

func BuildAgentMode(addr string, pInterval, sInterval string) {
	if len(addr) == 0 {
		log.Fatal("server addr must be set")
	}

	pDuration, err := util.ParseDuration(pInterval)
	if err != nil {
		logrus.Errorln("parse prober duration flag failed ", err)
		return
	}

	sDuration, err := util.ParseDuration(sInterval)
	if err != nil {
		logrus.Errorln("parse sync duration flag failed ", err)
		return
	}

	ctxAll, cancelAll := context.WithCancel(context.Background())

	cli := initRpcCli(ctxAll,addr)

	if err != nil {
		logrus.Errorln("agent get cli failed ", err)
		return
	}

	var g run.Group
	{
		// 定时拉取mesh poll
		manager := initTargetManager(pDuration, sDuration, cli)
		g.Add(func() error {
			manager.start(ctxAll)
			return nil
		}, func(err error) {
			cancelAll()
		})
	}

	g.Run()
}
