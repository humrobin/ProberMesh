package agent

import (
	"bytes"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"os/exec"
	"strings"
	"time"
)

var cmdLine = "ping -q -A -f -s 100 -W 1000 -c 50 %s"

func doICMP(jobs *proberJob) {
	for _, job := range jobs.targets {
		ctx, _ := context.WithTimeout(context.TODO(), time.Second*5)
		cmd := exec.CommandContext(ctx, "/bin/bash", "-c", fmt.Sprintf(cmdLine, job))

		var (
			stdout, stderr bytes.Buffer
			outString      string
		)

		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			logrus.Errorln("cmd run failed ", err)
			if strings.Contains(err.Error(), "killed") {
				outString = "killed"
			}

			outString = string(stderr.Bytes())
		}
		outString = string(stdout.Bytes())

		fmt.Println(outString)
	}
}
