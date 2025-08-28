package mining

import (
	"sync"
	"web3_gui/utils"
)

var bytePool = sync.Pool{
	New: func() interface{} {
		// b := make([]byte, 1024)
		buf := utils.NewBufferByte(0)
		return &buf
	},
}
