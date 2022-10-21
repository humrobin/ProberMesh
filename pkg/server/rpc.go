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

func (s *Server) Report(req pb.ReportReq, resp *string) error {
	hd.report(req.Region, req.IP)
	return nil
}

func (s *Server) GetTargetPool(req pb.TargetPoolReq, resp *pb.TargetPoolResp) error {
	tcs := tp.getPool(req.SourceRegion)
	resp.Targets = tcs
	return nil
}

func (s *Server) ProberResultReport(reqs []*pb.PorberResultReq, resp *string) error {
	aggregator.Enqueue(reqs)
	return nil
}

func startRpcServer(addr string) error {
	server := rpc.NewServer()
	err := server.Register(new(Server))
	if err != nil {
		return err
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	var mh codec.MsgpackHandle
	mh.MapType = reflect.TypeOf(map[string]interface{}(nil))

	go func() {
		for {
			// 从accept中拿到一个客户端的连接
			conn, err := l.Accept()
			if err != nil {
				logrus.Errorln("listen accept failed ", err)
				time.Sleep(100 * time.Millisecond)
				continue
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
	}()
	return nil
}
