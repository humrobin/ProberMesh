package server

import (
	"bufio"
	"github.com/sirupsen/logrus"
	"github.com/ugorji/go/codec"
	"io"
	"net"
	"net/rpc"
	"probermesh/pkg/pb"
	"reflect"
	"time"
)

type Server struct{}

func (s *Server) Report(req pb.PingReq, resp *string) error {
	msg := "success"
	resp = &msg
	return nil
}

func (s *Server) GetTargetPool(req pb.TargetPoolReq, resp *pb.TargetPoolResp) error {
	tcs := GetTP().GetPool(req.SourceRegion)
	resp.Targets = tcs
	return nil
}

func (s *Server) ProberResultReport(reqs []*pb.PorberResultReq, resp *string) error {
	aggregator.Enqueue(reqs)
	return nil
}

func startRpcServer(addr string, errCh chan error, quit chan struct{}) {
	server := rpc.NewServer()
	err := server.Register(new(Server))
	if err != nil {
		logrus.Errorln("init rpc server failed ", err)
		errCh <- err
		return
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		logrus.Errorln("net listen failed ", err)
		errCh <- err
		return
	}
	var mh codec.MsgpackHandle
	mh.MapType = reflect.TypeOf(map[string]interface{}(nil))

	go func() {
		<-quit
		l.Close()
	}()

	for {
		// 从accept中拿到一个客户端的连接
		conn, err := l.Accept()
		if err != nil {
			logrus.Errorln("listen accept failed ", err)
			time.Sleep(100 * time.Millisecond)
			errCh <- err
			return
		}

		// 用bufferio做io解析提速
		var bufConn = struct {
			io.Closer
			*bufio.Reader
			*bufio.Writer
		}{
			conn,
			bufio.NewReader(conn),
			bufio.NewWriter(conn),
		}
		go server.ServeCodec(codec.MsgpackSpecRpc.ServerCodec(bufConn, &mh))
	}
}
