package utils

import (
	"fmt"
	"testing"
)

func TestSyncList(t *testing.T) {
	// syncListExample()
}
func syncListExample() {
	slist := NewSyncList()
	slist.Add(1)
	slist.Add(2)
	slist.Add(3)

	nums := slist.GetAll()
	fmt.Println("Sync list:", nums)
	temp := nums[:1]
	nums = append(temp, nums[2:]...)
	fmt.Println("Sync list:", nums)

	nums = slist.GetAll()
	fmt.Println("Sync list:", nums)

	slist.Remove(1)
	nums = slist.GetAll()
	fmt.Println("Sync list:", nums)
	slist.Remove(1)
	nums = slist.GetAll()
	fmt.Println("Sync list:", nums)
	slist.Remove(0)
	slist.Remove(0)
	nums = slist.GetAll()
	fmt.Println("Sync list:", nums)
}
