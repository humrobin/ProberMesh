package agent

import (
	"fmt"
	"github.com/toolkits/pkg/net/gobrpc"
)

type proberJob struct {
	proberType   string
	targets      []string
	sourceRegion string
	targetRegion string
	r            *gobrpc.RPCClient
}

func (p *proberJob) run() {
	switch p.proberType {
	case "http":
		fmt.Println("http探测 ", p.proberType, p.sourceRegion, p.targetRegion)
	case "icmp":
		doICMP(p)
		fmt.Println("icmp探测 ", p.proberType, p.sourceRegion, p.targetRegion)
	}
}
