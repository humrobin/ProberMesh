package pb

import "probermesh/config"

type ReportReq struct {
	IP     string `json:"ip"`
	Region string `json:"region"`
}

type TargetPoolReq struct {
	SourceRegion string
}

type TargetPoolResp struct {
	Targets map[string][]*config.ProberConfig
}

type PorberResultReq struct {
	// 探测类型
	ProberType string

	// 探测地址
	ProberTarget string

	// 本地ip
	LocalIP string

	// 原region
	SourceRegion string

	// 目的region
	TargetRegion string

	// 探测是否成功
	ProberSuccess bool

	// icmp的字段 resolve setup rtt
	ICMPDurations map[string]float64

	// http字段
	HTTPDurations map[string]float64
}


