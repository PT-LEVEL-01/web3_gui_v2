//go:build ios
// +build ios

package hardware

// 是否满足最低硬件要求,minCpuCores:最低Cpu核心数, minFreeMem:最低可用内存MB, minFreeDisk:最低可用磁盘MB
func CheckHardwareRequirement(minCpuCores, minFreeMem, minFreeDisk, minNetworkBandwidth uint32, peerIps []string) error {
	return nil
}
