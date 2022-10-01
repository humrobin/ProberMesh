package server

import (
	"context"
	"probermesh/config"
)

type targetsPool struct {
	pool map[string][]*config.ProberConfig
	cfg  *config.ProberMeshConfig

	done context.Context
}

var tp *targetsPool

func GetTP() *targetsPool {
	return tp
}

func initTargetsPool(ctx context.Context, cfg *config.ProberMeshConfig) *targetsPool {
	tp = &targetsPool{
		pool: make(map[string][]*config.ProberConfig),
		cfg:  cfg,
		done: ctx,
	}
	return tp
}

func (t *targetsPool) start() {
	for _, pc := range t.cfg.ProberConfigs {
		if _, ok := t.pool[pc.Region]; !ok {
			t.pool[pc.Region] = []*config.ProberConfig{pc}
		} else {
			t.pool[pc.Region] = append(t.pool[pc.Region], pc)
		}
	}

	<-t.done.Done()
	return
}

func (t *targetsPool) GetPool(sourceRegion string) map[string][]*config.ProberConfig {
	pcs := make(map[string][]*config.ProberConfig)

	for region, pc := range t.pool {
		if region != sourceRegion {
			pcs[region] = pc
		}
	}
	return pcs
}
