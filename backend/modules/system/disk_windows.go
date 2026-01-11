//go:build windows

package system

import (
	"os/exec"
	"strconv"
	"strings"
)

// DiskInfo 磁盘信息
type DiskInfo struct {
	Total   uint64
	Used    uint64
	Percent float64
}

// GetDiskInfo 获取磁盘信息 (Windows)
func GetDiskInfo(path string) DiskInfo {
	info := DiskInfo{}

	// 使用 wmic 获取 C: 盘信息
	out, err := exec.Command("wmic", "logicaldisk", "where", "DeviceID='C:'", "get", "Size,FreeSpace").Output()
	if err != nil {
		return info
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 2 {
			free, _ := strconv.ParseUint(fields[0], 10, 64)
			total, _ := strconv.ParseUint(fields[1], 10, 64)
			if total > 0 {
				info.Total = total
				info.Used = total - free
				info.Percent = float64(info.Used) / float64(total) * 100
			}
		}
	}

	return info
}
