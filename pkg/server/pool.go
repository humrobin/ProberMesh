package server

import (
	"context"
	"github.com/sirupsen/logrus"
	"probermesh/config"
	"time"
)

type targetsPool struct {
	pool map[string]map[string]*config.ProberConfig
	cfg  *config.ProberMeshConfig

	done  context.Context
	ready chan struct{}
}

type discoveryType string

const (
	StaticDiscovery  discoveryType = "static"
	DynamicDiscovery discoveryType = "dynamic"
)

var tp *targetsPool

func newTargetsPool(
	ctx context.Context,
	cfg *config.ProberMeshConfig,
	ready chan struct{},
	dt string,
) *targetsPool {
	tp = &targetsPool{
		pool:  make(map[string]map[string]*config.ProberConfig),
		cfg:   cfg,
		done:  ctx,
		ready: ready,
	}

	switch discoveryType(dt) {
	case DynamicDiscovery:
		logrus.Warnln("server use dynamic discovery type to find icmp targets by agent report")
		go tp.updatePool()
	case StaticDiscovery:
		logrus.Warnln("server use static discovery type to find icmp targets by config")
	}
	return tp
}

func (t *targetsPool) start() {
	for _, pc := range t.cfg.ProberConfigs {
		/*
			{
				cn-shanghai: {
					icmp: icmpConfig,
					http: httpConfig,
				}
			}
		*/
		if pcm, ok := t.pool[pc.Region]; ok {
			pcm[pc.ProberType] = pc
		} else {
			t.pool[pc.Region] = map[string]*config.ProberConfig{pc.ProberType: pc}
		}
	}

	<-t.done.Done()
	return
}

func (t *targetsPool) updatePool() {
	// 每分钟更新次pool值(根据agent上报)
	ticker := time.NewTicker(time.Duration(1) * time.Minute)
	defer ticker.Stop()

	do := func() {
		var updateKey = "icmp"
		for region, ips := range getDiscoverPool() {
			updatePC := &config.ProberConfig{
				ProberType: updateKey,
				Region:     region,
				Targets:    ips,
			}

			// 如果使用自动发现方式，则会覆盖掉配置中指定的同regon下的icmp targets节点
			// 全部使用agent上报的ip进行探测
			if pm, ok := t.pool[region]; ok {
				pm[updateKey] = updatePC
			} else {
				t.pool[region] = map[string]*config.ProberConfig{updateKey: updatePC}
			}
		}
	}

	<-t.ready
	do()
	for {
		select {
		case <-t.done.Done():
			return
		case <-ticker.C:
			do()
		}
	}
}

func (t *targetsPool) GetPool(sourceRegion string) map[string][]*config.ProberConfig {
	pcs := make(map[string][]*config.ProberConfig)
	for region, pcm := range t.pool {
		if region != sourceRegion {
			var ps []*config.ProberConfig
			for _, pc := range pcm {
				ps = append(ps, pc)
			}
			pcs[region] = ps
		}
	}
	return pcs
}
