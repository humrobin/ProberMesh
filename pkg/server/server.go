package server

import (
	"context"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"probermesh/config"
	"probermesh/pkg/util"
	"syscall"
)

func BuildServerMode(configPath string) {
	if err := config.InitConfig(configPath); err != nil {
		logrus.Fatal("server parse config failed ", err)
	}

	cfg := config.Get()

	ctxAll, cancelAll := context.WithCancel(context.Background())
	ctxAll = ctxAll

	var g run.Group
	{
		// rpc server
		errCh := make(chan error, 1)
		quit := make(chan struct{})
		g.Add(func() error {
			startRpcServer(cfg.RPCListenAddr, errCh, quit)
			return <-errCh
		}, func(err error) {
			cancelAll()
			close(quit)
			logrus.Warnln("rpc server over")
		})
	}

	{
		// 初始化targetsPool
		g.Add(func() error {
			initTargetsPool(ctxAll, cfg).start()
			return nil
		}, func(err error) {
			cancelAll()
		})
	}

	{
		// aggregation
		g.Add(func() error {
			aggD, err := util.ParseDuration(cfg.AggregationInterval)
			if err != nil {
				logrus.Errorln("agg interval parse failed ", err)
				return err
			}
			NewAggregator(ctxAll, aggD).startAggregation()
			return nil
		}, func(err error) {
			cancelAll()
		})
	}

	{
		// http server
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		svc := http.Server{
			Addr:    cfg.HTTPListenAddr,
			Handler: mux,
		}

		errCh := make(chan error)
		go func() {
			errCh <- svc.ListenAndServe()
		}()

		g.Add(func() error {
			select {
			case <-errCh:
			case <-ctxAll.Done():
			}
			return svc.Shutdown(context.TODO())
		}, func(err error) {
			cancelAll()
		})
	}

	{
		// 信号管理
		term := make(chan os.Signal, 1)
		signal.Notify(term, os.Interrupt, syscall.SIGTERM)
		cancel := make(chan struct{})

		g.Add(func() error {
			select {
			case <-term:
				logrus.Warnln("优雅关闭ing...")
				cancelAll()
				return nil
			case <-cancel:
				return nil
			}
		}, func(err error) {
			close(cancel)
			logrus.Warnln("signal controller over")
		})
	}

	g.Run()
}
