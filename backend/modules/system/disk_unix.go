//go:build !windows

package system

import "syscall"

// DiskInfo 磁盘信息
type DiskInfo struct {
	Total   uint64
	Used    uint64
	Percent float64
}

// GetDiskInfo 获取磁盘信息 (Unix/Linux/macOS)
func GetDiskInfo(path string) DiskInfo {
	var stat syscall.Statfs_t
	info := DiskInfo{}

	if err := syscall.Statfs(path, &stat); err == nil {
		info.Total = stat.Blocks * uint64(stat.Bsize)
		info.Used = (stat.Blocks - stat.Bfree) * uint64(stat.Bsize)
		if info.Total > 0 {
			info.Percent = float64(info.Used) / float64(info.Total) * 100
		}
	}

	return info
}
