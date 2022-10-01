package main

import (
	"flag"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
	"probermesh/pkg/agent"
	"probermesh/pkg/server"
	"probermesh/pkg/version"
	"time"
)

const projectName = "ProberMesh"

var (
	configPath string
	mode       string
	serverAddr string
	h          bool
	v          bool
)

func init() {
	initArgs()
	initLog()
}

func initLog() {
	var (
		// 日志级别
		logLevel = logrus.DebugLevel
		// 日志格式
		format = &logrus.TextFormatter{TimestampFormat: "2006-01-02 15:04:05"}

		// 日志文件根目录
		logDir  = "/logs/"
		logName = projectName
		logPath = path.Join(logDir, logName+".log")
	)

	// prod 环境提高日志等级
	logLevel = logrus.WarnLevel

	// 持久化日志
	if _, err := os.Stat(logDir); err != nil && os.IsNotExist(err) {
		if err := os.Mkdir(logDir, 0666); err != nil {
			logrus.WithFields(logrus.Fields{
				"err": err,
			}).Fatalln("can not mkdir logs")
		}
	}

	// 创建log文件
	file, err := os.OpenFile(
		logPath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0666,
	)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Fatalln("can not find logs dir or file")
	}
	writers := []io.Writer{file, os.Stdout}
	multiWriters := io.MultiWriter(writers...)

	// 同时记录gin的日志
	//gin.DefaultWriter = multiWriters
	// 同时写文件和屏幕
	logrus.SetOutput(multiWriters)

	// 配置log分割
	logf, err := rotatelogs.New(
		// 切割的日志名称
		path.Join(logDir, logName)+"-%Y%m%d.log",
		// 日志软链
		rotatelogs.WithLinkName(logPath),
		// 日志最大保存时长
		rotatelogs.WithMaxAge(time.Duration(7*24)*time.Hour),
		// 日志切割时长
		rotatelogs.WithRotationTime(time.Duration(24)*time.Hour),
	)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Fatalln("can not new rotatelog")
	}

	// 添加logrus钩子
	hook := lfshook.NewHook(
		lfshook.WriterMap{
			logrus.InfoLevel:  logf,
			logrus.FatalLevel: logf,
			logrus.DebugLevel: logf,
			logrus.WarnLevel:  logf,
			logrus.ErrorLevel: logf,
			logrus.PanicLevel: logf,
		},
		format,
	)
	logrus.AddHook(hook)

	// 设置日志记录级别
	logrus.SetLevel(logLevel)
	// 输出日志中添加文件名和方法信息
	logrus.SetReportCaller(true)
	// 设置日志格式
	logrus.SetFormatter(format)
}

func initArgs() {
	flag.StringVar(&configPath, "config.file", "prober_mesh.yaml", "指定config path")
	flag.StringVar(&mode, "mode", "server", "服务模式, agent/server")
	flag.StringVar(&serverAddr, "rpc.server.addr", "", "server rpc地址")
	flag.BoolVar(&v, "v", false, "版本信息")
	flag.BoolVar(&h, "h", false, "帮助信息")
	flag.Parse()

	if v {
		logrus.WithField("version", version.Version).Println("version")
		os.Exit(0)
	}

	if h {
		flag.Usage()
		os.Exit(0)
	}
}

func main() {
	switch mode {
	case "agent":
		logrus.Warnln("build agent mode")
		agent.BuildAgentMode(serverAddr)
	case "server":
		logrus.Warnln("build server mode")
		server.BuildServerMode(configPath)
	default:
		logrus.Fatal("mode must in agent/server")
	}
}
