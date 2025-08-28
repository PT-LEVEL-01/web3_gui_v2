package utils

import (
	"fmt"
	"testing"
)

func TestFile(t *testing.T) {
	//file_example()
}

func file_example() {
	//filePath := "D:/hny_im.zip"
	filePath := "D:/迅雷下载/BB1mE6NU.png"
	mimeType, _, err := FileContentType(filePath)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	fmt.Println(mimeType)

	ok, mimeType, _, err := FileContentTypeIsImage(filePath)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	fmt.Println(ok, mimeType)
}
