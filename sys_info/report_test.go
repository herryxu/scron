package sys_info

import (
	"fmt"
	"github.com/henryxu/tools/common"
	"testing"
)

func TestReportEnv(t *testing.T) {
	fmt.Println(common.RunMode)
	sysInfo := &SysInfo{
		Ip:     "",
		Cpu:    0,
		Memory: 0,
		Dick:   0,
	}
	sysInfo.ReportEnv("alpha")
	fmt.Println(common.RunMode)
}
