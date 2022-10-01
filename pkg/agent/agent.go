package agent

import (
	"context"
	"github.com/oklog/run"
	"github.com/sirupsen/logrus"
	"log"
)

func BuildAgentMode(addr string) {
	if len(addr) == 0 {
		log.Fatal("server addr must be set")
	}

	cli := initRpcCli(addr)

	r, err := cli.getCli()
	if err != nil {
		logrus.Errorln("agent get cli failed ", err)
		return
	}

	// 测试rpc cli
	//var msg string
	//err = r.Call("Server.Report", pb.PingReq{
	//	IP:     getLocalIP(),
	//	Region: "cn-shanghai",
	//}, &msg)
	//if err != nil {
	//	logrus.Errorln("agent call cli failed ", err)
	//	return
	//}
	//fmt.Println("cli 成功 ", msg)

	ctxAll, cancelAll := context.WithCancel(context.Background())

	var g run.Group
	{
		// 定时拉取mesh poll
		manager := initTargetManager(10, r)
		g.Add(func() error {
			manager.start(ctxAll)
			return nil
		}, func(err error) {
			cancelAll()
		})
	}

	g.Run()
}
