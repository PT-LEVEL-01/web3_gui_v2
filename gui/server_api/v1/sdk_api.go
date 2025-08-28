package server_api

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"github.com/kbinani/screenshot"
	"github.com/nickalie/go-webpbin"
	"github.com/wailsapp/wails/v3/pkg/application"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
	imconfig "web3_gui/config"
	"web3_gui/im/im"
	imModel "web3_gui/im/model"
	"web3_gui/libp2parea/v1/cake/update_version"
	"web3_gui/libp2parea/v1/sdk/config"
	"web3_gui/libp2parea/v1/sdk/jsonrpc2/model"
	"web3_gui/utils"
)

// App struct
type SdkApi struct {
	App      *application.App
	Ctx      context.Context
	dirIndex *imModel.DirectoryIndex //
}

// NewApp creates a new App application struct
func NewSDKApi() *SdkApi {
	return &SdkApi{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *SdkApi) Startup(ctx context.Context) {
	a.Ctx = ctx
}

/*
查询版本号
*/
func (a *SdkApi) GetVersion() (string, int) {
	return imconfig.GetVersion()
}

/*
检查并更新版本
@return    bool      是否有新版本
@return    bool      是否需要重启
@return    string    新版本名称
@return    int       新版本编号
*/
func (a *SdkApi) CheckUpdateVersion() VersionUpdateInfo {
	isLatest, fileName, versionName, localIndex, _, err := im.UpdateVersionModule.CheckLatestVersion()
	if err != nil {
		utils.Log.Info().Msgf("检查版本错误:%s", err.Error())
		return VersionUpdateInfo{IsNew: false, IsRester: false, VersionName: "", VersionIndex: 0, Error: err.Error()}
	}
	if isLatest {
		utils.Log.Info().Msgf("是最新版本")
		//判断是否需要重启
		if imconfig.VersionIndex >= localIndex {
			return VersionUpdateInfo{IsNew: false, IsRester: false, VersionName: versionName, VersionIndex: localIndex, Error: ""}
		}
		return VersionUpdateInfo{IsNew: true, IsRester: true, VersionName: versionName, VersionIndex: localIndex, Error: ""}
	}
	//检查本地文件是否已经存在
	filePath := filepath.Join(runtime.GOOS, fileName)
	exist, err := utils.PathExists(filePath)
	if err != nil {
		return VersionUpdateInfo{IsNew: false, IsRester: false, VersionName: "", VersionIndex: 0, Error: err.Error()}
	}
	utils.Log.Info().Msgf("是否存在新版本文件：%s %t", fileName, exist)
	if exist {
		return VersionUpdateInfo{IsNew: true, IsRester: true, VersionName: versionName, VersionIndex: localIndex, Error: ""}
	}
	//检查本地正在下载的临时文件是否已经存在
	filePath = filepath.Join(runtime.GOOS, fileName+update_version.TempFileSuffix)
	exist, err = utils.PathExists(filePath)
	if err != nil {
		return VersionUpdateInfo{IsNew: false, IsRester: false, VersionName: "", VersionIndex: 0, Error: ""}
	}
	utils.Log.Info().Msgf("是否存在临时文件：%s %t", fileName, exist)
	if exist {
		return VersionUpdateInfo{IsNew: true, IsRester: true, VersionName: versionName, VersionIndex: localIndex, Error: ""}
	}
	//开始异步下载文件
	go im.UpdateVersionModule.GetVersionFile(fileName)
	return VersionUpdateInfo{IsNew: true, IsRester: false, VersionName: versionName, VersionIndex: localIndex, Error: ""}
}

type VersionUpdateInfo struct {
	IsNew        bool   //是否有新版本
	IsRester     bool   //是否需要重启
	VersionName  string //新版本名称
	VersionIndex int    //新版本编号
	Error        string //
}

/*
检查密码是否正确
*/
func (a *SdkApi) CheckPassword(pwd string) bool {
	return true
}

/*
获取本节点网络地址和连接的其他节点信息
*/
func (a *SdkApi) GetNetwork() *NetWorkinfo {
	addrs := make([]string, 0)
	for _, one := range *im.Area.GetNetworkInfo() {
		addrs = append(addrs, one.B58String())
	}

	netinfo := NetWorkinfo{
		NetAddr:   im.Area.GetNetId().B58String(),
		Issuper:   im.Area.NodeManager.NodeSelf.GetIsSuper(),
		WebAddr:   config.WebAddr + ":" + strconv.Itoa(int(config.WebPort)),
		TCPAddr:   im.Area.NodeManager.NodeSelf.Addr + ":" + strconv.Itoa(int(im.Area.NodeManager.NodeSelf.TcpPort)), //
		LogicAddr: addrs,
	}
	return &netinfo
}

type NetWorkinfo struct {
	NetAddr   string   //本节点地址
	Issuper   bool     //是否超级节点
	WebAddr   string   //
	TCPAddr   string   //
	LogicAddr []string //连接的其他节点地址
}

/*
更新并重启
*/
func (a *SdkApi) UpdateRestar() int {
	fileName := os.Args[0]
	cmd := exec.Command(imconfig.UpdateProcName, fileName)
	err := cmd.Start()
	if err != nil {
		utils.Log.Error().Msgf("更新并重启错误:%s", err.Error())
		return model.Nomarl
	}
	a.App.Quit()
	//wailsruntime.Quit(context.Background())
	return model.Success
}

/*
屏幕快照
@isMini    bool    窗口是否最小化
*/
func (a *SdkApi) GetScreenShot(isMini bool) map[string]interface{} {
	if isMini {
		//窗口最小化
		a.App.Hide()
		//wailsruntime.WindowMinimise(a.Ctx)
		time.Sleep(time.Second)
	}
	resultMap := make(map[string]interface{})
	base64Str, err := screenShotV2()
	if err != nil {
		ERR := utils.NewErrorSysSelf(err)
		resultMap["code"] = ERR.Code
		resultMap["error"] = ERR.Msg
		return resultMap
	}
	if isMini {
		//窗口取消最小化
		a.App.Show()
		//wailsruntime.WindowUnminimise(a.Ctx)
	}
	ERR := utils.NewErrorSuccess()
	resultMap["info"] = base64Str
	resultMap["code"] = ERR.Code
	return resultMap
}

/*
截屏
*/
func screenShotV2() (string, error) {
	//捕获每个显示。
	n := screenshot.NumActiveDisplays()
	if n <= 0 {
		//panic("Active display not found")
		return "", errors.New("Active display not found")
	}
	var all image.Rectangle = image.Rect(0, 0, 0, 0)
	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		all = bounds.Union(all)
		_, err := screenshot.CaptureRect(bounds)
		if err != nil {
			return "", err
		}
	}
	img, err := screenshot.Capture(all.Min.X, all.Min.Y, all.Dx(), all.Dy())
	if err != nil {
		//panic(err)
		return "", err
	}

	base64Str := ""
	if false {
		buf := bytes.NewBuffer(nil)
		err = png.Encode(buf, img)
		if err != nil {
			//log.Error(err)
			return "", err
		}
		// 将字节切片转换为Base64字符串
		base64Str = base64.StdEncoding.EncodeToString(buf.Bytes())
		return "data:image/png;base64," + base64Str, nil
	} else {
		//转为webp
		minimum := uint(1)
		maximum := uint(100)
		trySize := uint(50)
		var webpBytes []byte
		for {
			webpBuf := bytes.NewBuffer(nil)
			err = webpbin.NewCWebP().Quality(trySize).InputImage(img).Output(webpBuf).Run()
			//err = webpbin.Encode(webpBuf, img)
			//err = webp.Encode(webpBuf, img, &encoder.Options{Quality: trySize})
			//webpBytes, err = webp.EncodeRGBA(img, trySize)
			if err != nil {
				utils.Log.Info().Msgf("序列化报错:%s", err.Error())
				if webpBytes != nil && len(webpBytes) > 0 {
					break
				}
				return "", err
			}
			webpBytes = webpBuf.Bytes()
			utils.Log.Info().Msgf("本次大小:%d %d %d %d", len(webpBytes), minimum, maximum, trySize)

			//break
			//太大了
			if utils.Byte(len(webpBytes)) > imconfig.ScreenShotMaxLength {
				oldTrySize := trySize
				trySize = (oldTrySize - minimum) / 2
				if oldTrySize == trySize {
					break
				}
				maximum = oldTrySize
			} else if utils.Byte(len(webpBytes)) < imconfig.ScreenShotMinLength {
				//太小了
				oldTrySize := trySize
				trySize = (maximum - oldTrySize) / 2
				if oldTrySize == trySize {
					break
				}
				minimum = oldTrySize
			} else {
				break
			}
		}
		utils.Log.Info().Msgf("最终大小:%d", len(webpBytes))
		// 将字节切片转换为Base64字符串
		base64Str = base64.StdEncoding.EncodeToString(webpBytes)
		return "data:image/webp;base64," + base64Str, nil
	}
}

//
//// save *image.RGBA to filePath with PNG format.
//func save(img *image.RGBA, filePath string) {
//	file, err := os.Create(filePath)
//	if err != nil {
//		panic(err)
//	}
//	defer file.Close()
//	err = png.Encode(file, img)
//	if err != nil {
//		panic(err)
//	}
//}
//
//func encode(img *image.RGBA) string {
//	//fSrc, err := os.Open("test.png")
//	//defer fSrc.Close()
//	//
//	//img, err = png.Decode(fSrc)
//	//if err != nil {
//	//	return nil, err
//	//}
//
//	// 这里的resImg是一个 image.Image 类型的变量
//	//buf := bytes
//	var buf bytes.Buffer
//	err := png.Encode(&buf, img)
//	if err != nil {
//		//log.Error(err)
//		return ""
//	}
//
//	//转为webp
//	webpBytes, err := webp.EncodeRGBA(img, 20)
//	if err != nil {
//		//log.Error(err)
//		return ""
//	}
//
//	fmt.Println("png", len(buf.Bytes()), "webp", len(webpBytes))
//
//	// 将字节切片转换为Base64字符串
//	base64Str := base64.StdEncoding.EncodeToString(webpBytes)
//	//fmt.Println("编码：", base64Str)
//	//screenBase64Str = base64Str
//	return base64Str
//}
//
///*
//截屏
//*/
//func screenShot() string {
//	// Capture each displays.
//	n := screenshot.NumActiveDisplays()
//	if n <= 0 {
//		panic("Active display not found")
//	}
//
//	var all image.Rectangle = image.Rect(0, 0, 0, 0)
//
//	for i := 0; i < n; i++ {
//		bounds := screenshot.GetDisplayBounds(i)
//		all = bounds.Union(all)
//
//		_, err := screenshot.CaptureRect(bounds)
//		if err != nil {
//			panic(err)
//		}
//		//fileName := fmt.Sprintf("%d_%dx%d.png", i, bounds.Dx(), bounds.Dy())
//		//save(img, fileName)
//		//
//		//fmt.Printf("#%d : %v \"%s\"\n", i, bounds, fileName)
//	}
//
//	// Capture all desktop region into an image.
//	fmt.Printf("%v\n", all)
//	img, err := screenshot.Capture(all.Min.X, all.Min.Y, all.Dx(), all.Dy())
//	if err != nil {
//		panic(err)
//	}
//	//save(img, "all.png")
//	return encode(img)
//}
//
//// ImageBytes2WebpBytes 将图片转为webp
//// inputFile 图片字节切片（仅限gif,jpeg,png格式）
//// outputFile webp图片字节切片
//// 图片质量
//func ImageBytes2WebpBytes(input []byte, quality float32) ([]byte, error) {
//
//	//解析图片
//	img, format, err := image.Decode(bytes.NewBuffer(input))
//	if err != nil {
//		log.Println("图片解析失败")
//		return nil, err
//	}
//
//	log.Println("原始图片格式：", format)
//
//	//转为webp
//	webpBytes, err := webp.EncodeRGBA(img, quality)
//
//	if err != nil {
//		log.Println("解析图片失败", err)
//		return nil, err
//	}
//
//	return webpBytes, nil
//}
//
//// Image2Webp 将图片转为webp
//// inputFile 图片路径（仅限gif,jpeg,png格式）
//// outputFile 图片输出路径
//// 图片质量
//func Image2Webp(inputFile string, outputFile string, quality float32) error {
//
//	// 读取文件
//	fileBytes, err := ioutil.ReadFile(inputFile)
//	if err != nil {
//		log.Println("读取文件失败:", err)
//		return err
//	}
//
//	webpBytes, err := ImageBytes2WebpBytes(fileBytes, quality)
//
//	if err != nil {
//		log.Println("解析图片失败", err)
//		return err
//	}
//
//	if err = ioutil.WriteFile(outputFile, webpBytes, 0666); err != nil {
//		log.Println("图片写入失败", err)
//		return err
//	}
//
//	originalSize := len(fileBytes)
//	webpSize := len(webpBytes)
//	log.Printf("原始大小:%d k,转换后大小:%d k,压缩比:%d %% \n", originalSize/1024, webpSize/1024, webpSize*100/originalSize)
//
//	return nil
//}
