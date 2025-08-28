package libp2parea

import (
	"sync"
)

type Global struct {
	lock  *sync.RWMutex    //
	areas map[string]*Node //key:string=域名称;value:*Area=域;
}

func NewGlobal() *Global {
	g := new(Global)
	g.lock = new(sync.RWMutex)
	g.areas = make(map[string]*Node)
	return g
}
