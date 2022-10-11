package agent

import (
	"context"
	"github.com/sirupsen/logrus"
	"probermesh/pkg/pb"
	"sync"
	"time"
)

type proberJob struct {
	proberType   string
	targets      []string
	sourceRegion string
	targetRegion string
	r            *rpcCli
}

func (p *proberJob) run() {
	ctx, _ := context.WithTimeout(context.TODO(), time.Duration(5*float64(time.Second)))

	switch p.proberType {
	case "http":
		p.dispatch(ctx, "http")
	case "icmp":
		p.dispatch(ctx, "icmp")
	}
}

func (p *proberJob) dispatch(ctx context.Context, pType string) {
	var (
		ptsChan = make(chan *pb.PorberResultReq, len(p.targets))
		pts     = make([]*pb.PorberResultReq, 0, len(p.targets))
		wg      sync.WaitGroup
	)

	go func() {
		wg.Wait()
		// TODO 报错 想关闭的chan中写入
		close(ptsChan)
	}()

	for _, target := range p.targets {
		wg.Add(1)
		go func(target string) {
			defer wg.Done()
			if pType == "icmp" {
				ptsChan <- probeICMP(ctx, target, p.sourceRegion, p.targetRegion)
			} else {
				ptsChan <- probeHTTP(ctx, target, p.sourceRegion, p.targetRegion)
			}
		}(target)
	}

	for pt := range ptsChan {
		pts = append(pts, pt)
	}

	// batch send
	go func(pts []*pb.PorberResultReq) {
		if err := p.r.Get().Call(
			"Server.ProberResultReport",
			pts,
			nil,
		); err != nil {
			logrus.Errorln("prober report failed ", err)
		}
	}(pts)
}
