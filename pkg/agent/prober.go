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
	pt := "icmp"

	if p.proberType == "http" {
		pt = "http"
	}
	p.dispatch(ctx, pt)
}

func (p *proberJob) dispatch(ctx context.Context, pType string) {
	var (
		ptsChan = make(chan *pb.PorberResultReq, len(p.targets))
		pts     = make([]*pb.PorberResultReq, 0, len(p.targets))
		wg      sync.WaitGroup
	)

	for _, target := range p.targets {
		wg.Add(1)
		go func(target string) {
			defer wg.Done()
			if pType == "icmp" {
				ptsChan <- probeICMP(ctx, target, p.sourceRegion, p.targetRegion)
			} else {
				ptsChan <- probeHTTP(ctx, target, p.sourceRegion, p.targetRegion)
			}
			time.Sleep(1 * time.Second)
		}(target)
	}

	go func() {
		// wg.wait() 放在 wg.add 后
		wg.Wait()
		close(ptsChan)
	}()

	for pt := range ptsChan {
		pts = append(pts, pt)
	}

	// batch send
	go func(pts []*pb.PorberResultReq) {
		if err := p.r.Call(
			"Server.ProberResultReport",
			pts,
			nil,
		); err != nil {
			logrus.Errorln("prober report failed ", err)
		}
	}(pts)
}
