package server

import (
	"context"
	"github.com/sirupsen/logrus"
	"probermesh/pkg/pb"
	"sync"
	"time"
)

const (
	defaultKeySeparator = "_"
)

type Aggregator struct {
	queue       [][]*pb.PorberResultReq
	aggInterval time.Duration

	cancel context.Context
	m      sync.Mutex
}

type aggProberResult struct {
	sourceRegion, targetRegion string // icmp 使用
	targetAddr                 string // http 使用
	batchCnt                   int64  // cnt算avg

	failedCnt int64
	phase     map[string]float64
}

var aggregator *Aggregator

func newAggregator(ctx context.Context, interval time.Duration) *Aggregator {
	aggregator = &Aggregator{
		queue:       make([][]*pb.PorberResultReq, 0),
		aggInterval: interval,
		cancel:      ctx,
	}
	return aggregator
}

func (a *Aggregator) Enqueue(reqs []*pb.PorberResultReq) {
	a.m.Lock()
	defer a.m.Unlock()
	a.queue = append(a.queue, reqs)
}

func (a *Aggregator) startAggregation() {
	ticker := time.NewTicker(a.aggInterval)
	defer ticker.Stop()

	a.agg()
	for {
		select {
		case <-a.cancel.Done():
			return
		case <-ticker.C:
			// 定时聚合所有数据
			a.agg()
		}
	}
}

func (a *Aggregator) agg() {
	if len(a.queue) == 0 {
		logrus.Warnln("current batch no agent report, continue...")
		return
	}
	logrus.Warnln("has batch report to agg ", len(a.queue))

	var (
		/*
			icmpAggMap = {
			"beijing->shanghai": []PorberResultReq
			"shanghai->beijing": []PorberResultReq
			}
		*/
		icmpAggMap = make(map[string]*aggProberResult)
		httpAggMap = make(map[string]*aggProberResult)
	)

	a.m.Lock()
	defer func() {
		a.m.Unlock()
		a.reset()
	}()

	for _, prs := range a.queue {
		for _, pr := range prs {
			var (
				containers map[string]*aggProberResult
				phase      map[string]float64
				key        string
			)
			if pr.ProberType == "http" {
				containers = httpAggMap
				phase = pr.HTTPDurations
				key = pr.SourceRegion + defaultKeySeparator + pr.ProberTarget
			} else {
				containers = icmpAggMap
				phase = pr.ICMPDurations
				key = pr.SourceRegion + defaultKeySeparator + pr.TargetRegion
			}

			if _, ok := containers[key]; !ok {
				containers[key] = &aggProberResult{
					sourceRegion: pr.SourceRegion,
					targetRegion: pr.TargetRegion,
					targetAddr:   pr.ProberTarget,
					phase:        make(map[string]float64),
				}
			}

			// 只有 ProberSuccess 成功的任务才 batchCnt++
			// 保证算agg时，分母一定为成功的job数
			// 防止: 成功4台,失败1台；算agg: 理想 total/4, 结果 total/5, 反而会拉低实际值
			container := containers[key]
			if pr.ProberSuccess {
				container.batchCnt++

				// 仅累加成功任务的phase
				for stage, val := range phase {
					container.phase[stage] += val
				}
				continue
			}

			// 失败任务 failedCnt自增
			container.failedCnt++
		}
	}

	a.dotHTTP(httpAggMap)
	a.dotICMP(icmpAggMap)
}

func (a *Aggregator) dotHTTP(http map[string]*aggProberResult) {
	for _, agg := range http {
		if agg.failedCnt > 0 {
			httpProberFailedGaugeVec.WithLabelValues(
				agg.sourceRegion,
				agg.targetAddr,
			).Set(float64(agg.failedCnt))
		}

		for stage, total := range agg.phase {
			// 每个 sR->tR 的每个stage的平均
			httpProberDurationGaugeVec.WithLabelValues(
				stage,
				agg.sourceRegion,
				agg.targetAddr,
			).Set(total / float64(agg.batchCnt))
		}
	}
}

func (a *Aggregator) dotICMP(icmp map[string]*aggProberResult) {
	for _, agg := range icmp {
		// 当前 r to r 存在探测失败任务
		if agg.failedCnt > 0 {
			icmpProberFailedGaugeVec.WithLabelValues(
				agg.sourceRegion,
				agg.targetRegion,
			).Set(float64(agg.failedCnt))
		}

		var icmpDurationsTotal float64
		for stage, total := range agg.phase {
			stageAgg := total / float64(agg.batchCnt)

			switch stage {
			case "loss":
				// 单独打点丢包率指标
				icmpProberPacketLossRateGaugeVec.WithLabelValues(
					agg.sourceRegion,
					agg.targetRegion,
				).Set(stageAgg)
			case "stddev":
				// 单独打点 stddev 抖动指标
				icmpProberJitterStdDevGaugeVec.WithLabelValues(
					agg.sourceRegion,
					agg.targetRegion,
				).Set(stageAgg)
			default:
				// 每个 sR->tR 的每个stage的平均
				icmpProberDurationGaugeVec.WithLabelValues(
					stage,
					agg.sourceRegion,
					agg.targetRegion,
				).Set(stageAgg)
				icmpDurationsTotal += total
			}
		}

		// 为 r->r 打点histogram
		icmpProberDurationHistogramVec.WithLabelValues(
			agg.sourceRegion,
			agg.targetRegion,
		).Observe(icmpDurationsTotal)
	}
}

func (a *Aggregator) reset() {
	a.queue = make([][]*pb.PorberResultReq, 0)
}
