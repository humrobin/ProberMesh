package pb

import "probermesh/config"

type PingReq struct {
	IP     string `json:"ip"`
	Region string `json:"region"`
}

type TargetPoolReq struct {
	SourceRegion string
}

type TargetPoolResp struct {
	Targets map[string][]*config.ProberConfig
}
