package utils

import (
	"syscall"
)

/*
获取磁盘空间大小以及可用空间大小
@return    uint64    磁盘总大小
@return    uint64    磁盘可用空间大小
@return    uint64    磁盘可用空间大小
*/
func GetDiskFreeSpace(dirPath string) (uint64, uint64, uint64, error) {
	var stat syscall.Statfs_t
	// wd, _ := os.Getwd()
	err := syscall.Statfs(dirPath, &stat)
	if err != nil {
		return 0, 0, 0, err
	}
	// log.Println(stat.Bavail * uint64(stat.Bsize))
	// log.Println(stat.Blocks * uint64(stat.Bsize))
	// log.Printf("%+v", stat)

	return stat.Blocks * uint64(stat.Bsize), stat.Bavail * uint64(stat.Bsize), stat.Bfree * uint64(stat.Bsize), nil
}
