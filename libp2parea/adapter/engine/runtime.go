package engine

import (
	"sync"
	"time"
	"web3_gui/utils"
)

var runtimeMap = new(sync.Map)

func InitRuntimeMap() {
	go func() {
		time.Sleep(time.Minute * 30)
		runtimeMap.Range(func(k, v interface{}) bool {
			t, ok := v.(*time.Time)
			if !ok {
				return false
			}
			utils.Log.Info().Msgf("RuntimeMapOne :%s %s", k.(string), t)
			return true
		})
	}()
}

func AddRuntime(file string, line int, goroutineId string) {
	// now := time.Now()
	// key := file + "_" + strconv.Itoa(line) + "_" + goroutineId
	// utils.Log.Info().Msgf("AddRuntime :%s", key)
	// runtimeMap.Store(key, &now)
}

func DelRuntime(file string, line int, goroutineId string) {
	// key := file + "_" + strconv.Itoa(line) + "_" + goroutineId
	// utils.Log.Info().Msgf("DelRuntime :%s", key)
	// runtimeMap.Delete(key)
}
