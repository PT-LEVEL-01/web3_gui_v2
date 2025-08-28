package mining

//
//import (
//	"fmt"
//	"testing"
//)
//
//func TestNodeManager(t *testing.T) {
//	ImitateWork(t)
//}
//
//func ImitateWork(t *testing.T) {
//	wml := NewWorkModeLock()
//	//设置初始导入
//	wml.GetImportBlockLock()
//	wml.BackImportBlockLockSuccess(1, 1)
//
//	//已经导入过的高度，不给签名
//	ok := wml.CheckSign(1, 1)
//	fmt.Println("第1次请求签名:", ok, "应该为:false")
//	//模拟不同组高度，相同块高度，不给签名
//	ok = wml.CheckSign(2, 1)
//	fmt.Println("第2次请求签名:", ok, "应该为:false")
//	ok = wml.CheckSign(1, 2)
//	fmt.Println("第3次请求签名:", ok, "应该为:true")
//	ok = wml.CheckSign(1, 3)
//	fmt.Println("第4次请求签名:", ok, "应该为:true")
//	ok = wml.CheckSign(1, 4)
//	fmt.Println("第5次请求签名:", ok, "应该为:true")
//	//模拟组高度不相同，块高度相同，不给签名
//	ok = wml.CheckSign(2, 4)
//	fmt.Println("第6次请求签名:", ok, "应该为:false")
//	ok = wml.CheckSign(2, 5)
//	fmt.Println("第7次请求签名:", ok, "应该为:true")
//	//模拟组高度不相同，相同块高度，不给签名
//	ok = wml.CheckSign(3, 5)
//	fmt.Println("第8次请求签名:", ok, "应该为:false")
//	ok = wml.CheckSign(3, 6)
//	fmt.Println("第9次请求签名:", ok, "应该为:true")
//	//模拟组高度太小，块高度+1，不给签名
//	ok = wml.CheckSign(2, 7)
//	fmt.Println("第10次请求签名:", ok, "应该为:false")
//}
