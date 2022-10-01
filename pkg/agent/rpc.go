package agent

import (
	"bufio"
	"github.com/sirupsen/logrus"
	"github.com/toolkits/pkg/net/gobrpc"
	"github.com/ugorji/go/codec"
	"io"
	"net"
	"net/rpc"
	"reflect"
	"time"
)

type rpcCli struct {
	cli        *gobrpc.RPCClient
	serverAddr string
}

func initRpcCli(addr string) *rpcCli {
	return &rpcCli{
		serverAddr: addr,
	}
}

func (r *rpcCli) getCli() (*gobrpc.RPCClient, error) {
	if r.cli != nil {
		return r.cli, nil
	}

	timeout := time.Duration(5) * time.Second
	conn, err := net.DialTimeout("tcp", r.serverAddr, timeout)
	if err != nil {
		logrus.Errorln("get cli failed ", err)
		return nil, err
	}

	var bufConn = struct {
		io.Closer
		*bufio.Reader
		*bufio.Writer
	}{
		conn,
		bufio.NewReader(conn),
		bufio.NewWriter(conn),
	}

	var mh codec.MsgpackHandle
	mh.MapType = reflect.TypeOf(map[string]interface{}(nil))

	rpcCodec := codec.MsgpackSpecRpc.ClientCodec(bufConn, &mh)
	r.cli = gobrpc.NewRPCClient(
		r.serverAddr,
		rpc.NewClientWithCodec(rpcCodec),
		timeout,
	)
	return r.cli, nil
}

func (r *rpcCli) closeCli() {
	if r.cli != nil {
		r.cli.Close()
		r.cli = nil
	}
}
