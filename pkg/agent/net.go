package agent

import "net"

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
