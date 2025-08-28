package utils

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

// go tool pprof -http 127.0.0.1:8081 .\mem.prof
func PprofMem(interval time.Duration) {
	go func() {

		memf, err := os.Create("mem.prof")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Can not create mem profile output file: %s", err)
			return
		}

		cpuf, err := os.Create("cpu.prof")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Can not create cpu profile output file: %s", err)
			return
		}

		err = pprof.StartCPUProfile(cpuf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s", err)
			return
		}
		//	f, err := os.OpenFile("mem.prof", os.O_RDWR|os.O_CREATE, 0644)
		//	if err != nil {
		//		log.Fatal(err)
		//	}
		//	defer f.Close()
		//	pprof.StartCPUProfile(f)
		//	defer pprof.StopCPUProfile()

		//	pprof.Lookup("")

		//	f, err := os.Create("mem.prof")
		//	if err != nil {
		//		fmt.Fprintf(os.Stderr, "Can not create mem profile output file: %s", err)
		//		return
		//	}
		//	if err = pprof.WriteHeapProfile(f); err != nil {
		//		fmt.Fprintf(os.Stderr, "Can not write %s: %s", *memProfile, err)
		//	}
		//	f.Close()

		// runtime.GOMAXPROCS(1)              // 限制 CPU 使用数，避免过载
		// runtime.SetMutexProfileFraction(1) // 开启对锁调用的跟踪
		// runtime.SetBlockProfileRate(1)     // 开启对阻塞操作的跟踪

		runtime.MemProfileRate = 512 * 1024 //512k

		// startMemProfile()
		//time.Sleep(timeout)
		//ticker := time.NewTicker(10 * time.Minute)
		for _ = range time.Tick(interval) {
			StopMemProfile(memf, cpuf)
		}
	}()
}

func StopMemProfile(memProfile, cpuProfile *os.File) {
	// f, err := os.Create(memProfile)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Can not create mem profile output file: %s", err)
	// 	return
	// }
	if err := pprof.WriteHeapProfile(memProfile); err != nil {
		fmt.Fprintf(os.Stderr, "Can not write %s: %s", memProfile.Name(), err)
	}
	memProfile.Close()

	pprof.StopCPUProfile()
	cpuProfile.Close()

	// f, err = os.Create(runtimeFile)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Can not create mem profile output file: %s", err)
	// 	return
	// }
	// p := pprof.Lookup("goroutine")
	// p.WriteTo(w, 1)

	// if err = pprof.WriteHeapProfile(f); err != nil {
	// 	fmt.Fprintf(os.Stderr, "Can not write %s: %s", memProfile, err)
	// }
	// f.Close()
}
