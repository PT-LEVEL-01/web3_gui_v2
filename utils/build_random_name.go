package utils

import (
	"bufio"
	"bytes"
	_ "embed"
	"math/rand"
	"sync"
	"time"
)

var Name_pro []string               //名称前缀 rune
var Name_suffix_short []string      //名称后缀（短）
var Name_suffix_long []string       //名称后缀（长）
var initRandomName = new(sync.Once) //

//go:embed name_adj.text
var nameAdjByte []byte

//go:embed name_noun.text
var nameNounByte []byte

/*
	加载名称前缀和后缀
*/
func InitBuildRandomName() {
	initRandomName.Do(func() {
		//加载名称前缀
		// file, _ := os.Open("name_adj.text")
		buf := bufio.NewReader(bytes.NewBuffer(nameAdjByte))
		for {
			line, _, err := buf.ReadLine()
			if err != nil {
				break
			}
			Name_pro = append(Name_pro, string(line))
		}
		// file.Close()
		//加载名称后缀
		// file, _ = os.Open("name_noun.text")
		buf = bufio.NewReader(bytes.NewBuffer(nameNounByte))
		for {
			line, _, err := buf.ReadLine()
			if err != nil {
				break
			}
			Name_suffix_long = append(Name_suffix_long, string(line))
			if len([]rune(string(line))) < 3 {
				Name_suffix_short = append(Name_suffix_short, string(line))
			}
		}
		// file.Close()
	})

}

/*
	随机构建一个名称
*/
func BuildName() string {
	InitBuildRandomName()
	rand.Seed(int64(time.Now().Nanosecond()))
	index_pro := rand.Intn(len(Name_pro))
	if len([]rune(Name_pro[index_pro])) > 3 {
		index_suf := rand.Intn(len(Name_suffix_short))
		return Name_pro[index_pro] + Name_suffix_short[index_suf]
	} else {
		index_suf := rand.Intn(len(Name_suffix_long))
		return Name_pro[index_pro] + Name_suffix_long[index_suf]
	}
}

/*
	随机构建一个性别
*/
func BuildGender() int {
	rand.Seed(int64(time.Now().Nanosecond()))
	return rand.Intn(2)
}
