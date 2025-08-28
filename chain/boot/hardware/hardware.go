//go:build !ios
// +build !ios

package hardware

import (
	"bytes"
	"fmt"
	"github.com/go-ping/ping"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"os"
	"runtime"
	"time"
	"web3_gui/libp2parea/adapter/engine"
)

// 是否满足最低硬件要求,minCpuCores:最低Cpu核心数, minFreeMem:最低可用内存MB, minFreeDisk:最低可用磁盘MB
func CheckHardwareRequirement(minCpuCores, minFreeMem, minFreeDisk, minNetworkBandwidth uint32, peerIps []string) error {
	buf := bytes.NewBuffer(nil)
	table := tablewriter.NewWriter(buf)
	table.SetRowLine(true)
	table.SetRowSeparator("-")
	table.SetHeader([]string{"硬件配置项", "最低硬件配置要求", "当前硬件配置"})
	defer func() {
		table.Render()
		engine.Log.Info("最低硬件配置检测 %s %s\n%s", runtime.GOOS, runtime.GOARCH, buf.String())
	}()

	//CPU
	if minCpuCores != 0 {
		cpuCores, err := cpu.Counts(false)
		if err != nil {
			return err
		}
		if cpuCores < int(minCpuCores) {
			return errors.Errorf("low cpu cores %d", cpuCores)
		}
		table.Append([]string{"CPU 核心数", fmt.Sprintf("%d", minCpuCores), fmt.Sprintf("%d", cpuCores)})
	} else {
		table.Append([]string{"CPU 核心数", "-", "-"})
	}

	//内存
	if minFreeMem != 0 {
		vmem, err := mem.VirtualMemory()
		if err != nil {
			return err
		}
		freeMemVal := vmem.Free / (1024 * 1024) //转为MB
		if freeMemVal < uint64(minFreeMem) {
			return errors.Errorf("low free memory %dMB", freeMemVal)
		}
		table.Append([]string{"内存(MB)", fmt.Sprintf("%d", minFreeMem), fmt.Sprintf("%d", freeMemVal)})
	} else {
		table.Append([]string{"内存(MB)", "-", "-"})
	}

	//磁盘
	if minFreeDisk != 0 {
		progPath, err := os.Getwd()
		if err != nil {
			return err
		}
		dstat, err := disk.Usage(progPath)
		if err != nil {
			return err
		}
		freeDiskVal := dstat.Free / (1024 * 1024) //转为MB
		if freeDiskVal < uint64(minFreeDisk) {
			return errors.Errorf("low free disk %dMB", freeDiskVal)
		}
		table.Append([]string{"磁盘(MB)", fmt.Sprintf("%d", minFreeDisk), fmt.Sprintf("%d", freeDiskVal)})
	} else {
		table.Append([]string{"磁盘(MB)", "-", "-"})
	}

	//测试带宽,有一个符合就通过
	if minNetworkBandwidth != 0 && len(peerIps) > 0 {
		networkBandwidthOk := false
		maxNetworkBandwidth := float64(0)
		for _, peerIp := range peerIps {
			if nb, err := testNetworkBandwidth(peerIp, minNetworkBandwidth, 5); err != nil {
				if maxNetworkBandwidth < nb {
					maxNetworkBandwidth = nb
				}
				engine.Log.Warn("Checked Low Network [%s] Bandwidth: %0.2fMB, Error: %v", peerIp, nb, err)
			} else {
				if maxNetworkBandwidth < nb {
					maxNetworkBandwidth = nb
				}
				networkBandwidthOk = true
				break
			}
		}

		table.Append([]string{"网络带宽(MB/s)", fmt.Sprintf("%d", minNetworkBandwidth), fmt.Sprintf("%0.2f", maxNetworkBandwidth)})
		if !networkBandwidthOk {
			return errors.New("low network bandwidth")
		}
	} else {
		table.Append([]string{"网络带宽(MB/s)", "-", "-"})
	}

	return nil
}

// 测试带宽,retry=重试次数
func testNetworkBandwidth(peerIp string, minNetworkBandwidth uint32, retry int) (float64, error) {
	pinger, err := ping.NewPinger(peerIp)
	if err != nil {
		return 0, err
	}
	pinger.Count = 5
	pinger.Timeout = time.Second * 5
	pinger.Size = 1024
	pinger.SetPrivileged(true)
	if err := pinger.Run(); err != nil {
		return 0, err
	}

	avgRtt := pinger.Statistics().AvgRtt
	if avgRtt == 0 {
		retry--
		if retry > 0 {
			return testNetworkBandwidth(peerIp, minNetworkBandwidth, retry)
		}
		return 0, errors.New("avg rtt 0")
	}

	networkBandwidth := (float64(pinger.Size*8*2) / avgRtt.Seconds()) / (1024 * 1024)
	if uint32(networkBandwidth) < minNetworkBandwidth {
		return 0, errors.Errorf("low network bandwidth %0.2fMB/s", networkBandwidth)
	}
	engine.Log.Info("Checked [To:%s] Network Bandwidth: %0.2fMB/s", peerIp, networkBandwidth)

	return networkBandwidth, nil
}
