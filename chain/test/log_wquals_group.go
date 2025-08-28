package main

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"strconv"
	"strings"

	"web3_gui/libp2parea/adapter/engine"
)

func main() {
	example()
}

func example() {
	logOneBs, err := ioutil.ReadFile("D:/test/test_local_cmd/peer5/logs/log.txt")
	if err != nil {
		panic(err.Error())
	}
	logTwoBs, err := ioutil.ReadFile("D:/test/test_local_cmd/peer3/logs/log.txt")
	if err != nil {
		panic(err.Error())
	}

	buferOne := bufio.NewReader(bytes.NewBuffer(logOneBs))

	buferTwo := bufio.NewReader(bytes.NewBuffer(logTwoBs))

	start := false
	showCount := 3
	showTotal := 0
	showIndex := 0
	witnessOneAddrs := [3]string{"", "", ""}
	witnessTwoAddrs := [3]string{"", "", ""}
	for {
		line, _, err := buferOne.ReadLine()
		if err != nil {
			break
		}
		if ParseLineHaveEnd(&line) {
			showTotal = 0
			witnessOneAddrs = [3]string{"", "", ""}
			witnessTwoAddrs = [3]string{"", "", ""}
			continue
		}
		n, witnessAddr := ParseLine(&line)
		if witnessAddr == "" {
			continue
		}
		if showTotal == 0 && start && showIndex+1 != n {
			continue
		}
		engine.Log.Info("分析到组高度:%d", n)
		if showIndex == n {
			witnessOneAddrs[showTotal] = witnessAddr
			engine.Log.Info("showTotal:%d", showTotal)
			engine.Log.Info("添加地址:%d %s", showTotal, witnessAddr)
			showTotal++
		} else {
			witnessOneAddrs[showTotal] = witnessAddr
			engine.Log.Info("添加地址:%d %s", showTotal, witnessAddr)
			showTotal = 1
			showIndex = n
			engine.Log.Info("showTotal:%d showIndex:%d", showTotal, showIndex)
			continue
		}
		if showTotal != showCount {
			engine.Log.Info("showTotal:%d showCount:%d", showTotal, showCount)
			continue
		}
		start = true
		engine.Log.Info("组内有:%d个见证人 %+v", showTotal, witnessOneAddrs)
		//开始读第二个文件
		showTwoTotal := 0
		for {
			line, _, err = buferTwo.ReadLine()
			if err != nil {
				break
			}
			n, witnessAddr = ParseLine(&line)
			if n < showIndex {
				showTwoTotal = 0
				continue
			}
			if n > showIndex {
				panic("第二个日志找到的组高度太大")
			}
			if n == showIndex {
				witnessTwoAddrs[showTwoTotal] = witnessAddr
				showTwoTotal++
			}
			if showTwoTotal == showCount {
				break
			}
		}

		engine.Log.Info("%+v %+v", witnessOneAddrs, witnessTwoAddrs)
		//找到了，开始对比
		for i, one := range witnessOneAddrs {
			if one == witnessTwoAddrs[i] {
				engine.Log.Info("见证人组相等:%d", showIndex)
			} else {
				engine.Log.Info("见证人组不相等:%d", showIndex)
				panic("见证人组不相等")
			}
		}

	}

}

func ParseLine(srcbs *[]byte) (int, string) {
	strs := strings.Split(string(*srcbs), "witness group:")
	if len(strs) <= 1 {
		return 0, ""
	}
	strs = strings.Split(strs[1], " ")
	n, err := strconv.Atoi(strs[0])
	if err != nil {
		panic(err.Error())
	}
	return n, strs[1]
}

func ParseLineHaveEnd(srcbs *[]byte) bool {

	return strings.Contains(string(*srcbs), "--------------")

}
