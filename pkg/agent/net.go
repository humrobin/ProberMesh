package agent

import (
	"context"
	"net"
	"os/exec"
	"time"
)

const (
	defaultRegion    = "cn-shanghai"
	defaultRegionCmd = "curl -s http://100.100.100.200/latest/meta-data/region-id"
)

func getLocalIP() string {
	var (
		addrs   []net.Addr
		addr    net.Addr
		ipNet   *net.IPNet
		isIpNet bool
		err     error
	)

	// 获取本机网卡的ip
	if addrs, err = net.InterfaceAddrs(); err != nil {
		return "0.0.0.0"
	}
	// 取第一个非lo的网卡IP
	for _, addr = range addrs {
		// ipv4  ipv6
		if ipNet, isIpNet = addr.(*net.IPNet); isIpNet && !ipNet.IP.IsLoopback() {
			// 跳过ipv6
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}
	return "0.0.0.0"
}

func getSelfRegion() string {
	ctx, _ := context.WithTimeout(context.TODO(), time.Duration(2)*time.Second)
	cmd := exec.CommandContext(
		ctx,
		"bash",
		"-c",
		defaultRegionCmd,
	)

	bs, err := cmd.CombinedOutput()
	if err != nil {
		return defaultRegion
	}
	return string(bs)
}
