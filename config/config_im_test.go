package config

import (
	"fmt"
	"testing"
)

func TestBuildImgBase64InfoV1(*testing.T) {
	//base64BuildAndParse()
}

/*
文件加密
*/
func base64BuildAndParse() {
	a := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAA8AAAAJCAYAAADtj3ZXAAAAAXNSR0IArs4c6QAAABlJREFUKFNjlJNX/s9AJmAc1UxayA3RAAMAlo4MYZjyjEIAAAAASUVORK5CYII="
	imgInfo, ERR := BuildImgBase64InfoV1(a)
	if ERR.CheckFail() {
		fmt.Println(ERR.String())
		return
	}
	fmt.Println(imgInfo)

	imgStr, ERR := ParseImgBase64InfoV1(imgInfo)
	if ERR.CheckFail() {
		fmt.Println(ERR.String())
		return
	}
	fmt.Println(imgStr)
}
