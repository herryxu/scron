package sys_info

import (
	"log"
	"net"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
)

type SysInfo struct {
	Ip     string
	Cpu    float64
	Memory float64
	Dick   float64
}

func getCpuPercent() float64 {
	percent, _ := cpu.Percent(time.Second, false)
	return percent[0]
}

func getMemPercent() float64 {
	memInfo, _ := mem.VirtualMemory()
	return memInfo.UsedPercent
}

func getDiskPercent() float64 {
	parts, _ := disk.Partitions(true)
	diskInfo, _ := disk.Usage(parts[0].Mountpoint)
	return diskInfo.UsedPercent
}
func getLocalIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Println("getLocalIP:", err)
		return ""
	}
	for _, iface := range ifaces {
		if iface.Name == "eth0" {
			addrs, err := iface.Addrs()
			if err != nil {
				log.Println("getLocalIP:", err)
				continue
			}
			for _, addr := range addrs {
				ipnet, ok := addr.(*net.IPNet)
				if ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						log.Println("getLocalIP IPv4 address: ", ipnet.IP.String())
						return ipnet.IP.String()
					} else {
						log.Println("getLocalIP IPv6 address: ", ipnet.IP.String())
					}
				}
			}
		}
	}
	return ""
}

func GetSysInfo() *SysInfo {
	return &SysInfo{
		Ip:     getLocalIP(),
		Cpu:    getCpuPercent(),
		Memory: getMemPercent(),
		Dick:   getDiskPercent(),
	}
}
