package main

import (
	"embed"
	_ "embed"
	"encoding/json"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"os"
	"web3_gui/config"
	"web3_gui/gui/server_api"
	"web3_gui/gui/tray"
	"web3_gui/im/boot"
	"web3_gui/utils"
)

// Wails使用Go的“嵌入”包将前端文件嵌入到二进制文件中。frontend/dist文件夹中的任何文件都将嵌入二进制文件中，并可供前端使用。
// See https://pkg.go.dev/embed for more information.

//go:embed frontend/dist
var assets embed.FS

func main() {
	defer utils.PrintPanicStack(nil)

	server_api.InitConfig()

	//fileName := "pid.json"
	//exist, err := CheckAndShowGui(fileName)
	//if err != nil {
	//	panic(err)
	//}
	//if exist {
	//	return
	//}

	StartModuls()

	sdkApi := server_api.NewSDKApi()
	sdkApi.Load()

	//
	//wails3版本文档 https://v3alpha.wails.io/
	app := application.New(application.Options{
		Name:        config.GUITitle,
		Description: "A demo of using raw HTML & CSS",
		Services: []application.Service{
			application.NewService(sdkApi),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})
	sdkApi.App = app
	tray.InitTray(app)

	//创建一个窗口
	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  config.GUITitle,
		Width:  1024,
		Height: 768,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
		DevToolsEnabled:  true,
		//ShouldClose:       tray.MastertWindowCloseCallback, //关闭窗口回调
		EnableDragAndDrop: true, //允许拖拽文件到窗口内
	})
	//给窗口添加拖拽文件事件
	window.OnWindowEvent(events.Common.WindowFilesDropped, func(event *application.WindowEvent) {
		//files := event.Context().DroppedFiles()
		//app.EmitEvent("dragfiles", files)
	})

	// Register a hook to hide the window when the window is closing
	window.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		// Hide the window
		//window.Hide()
		if !tray.MastertWindowCloseCallback(window) {
			// Cancel the event so it doesn't get destroyed
			e.Cancel()
		}
	})

	//运行应用程序。此操作将一直阻塞，直到应用程序退出为止。
	err := app.Run()
	if err != nil {
		utils.Log.Error().Err(err).Send()
	}
}

/*
启动其他模块
*/
func StartModuls() {
	go boot.StartRPCModule()
}

func CheckAndShowGui(fileName string) (bool, error) {
	exist, err := CheckGuiExists(fileName)
	if err != nil {
		return false, err
	}
	//程序未运行
	if !exist {
		err = SavePidFile(fileName)
		if err != nil {
			return false, err
		}
		return false, nil
	}
	//程序已经在运行，
	return true, nil
}

/*
保存pid信息到本地文件
*/
func SavePidFile(fileName string) error {
	//localPid := syscall.Getpid()
	//name, err := robotgo.FindName(localPid)
	//if err != nil {
	//	return err
	//}
	//fmt.Println("进程名称2", name)
	pIdInfo := ProcessIdInfo{
		Abs: os.Args[0],
		//Name: name,
		//PID:  localPid,
	}
	return utils.SaveJsonFile(fileName, pIdInfo)
}

/*
检查程序是否正在运行，如果正在运行，则窗口显示在最前面
@return    bool    是否存在
@return    error    错误
*/
func CheckGuiExists(fileName string) (bool, error) {
	exist, err := utils.PathExists(fileName)
	if err != nil {
		return false, err
	}
	//文件不存在，直接返回
	if !exist {
		return exist, nil
	}
	//文件存在，则解析文件
	bs, err := os.ReadFile(fileName)
	if err != nil {
		return false, err
	}
	pIdInfo := new(ProcessIdInfo)
	err = json.Unmarshal(bs, pIdInfo)
	if err != nil {
		return false, err
	}
	//判断进程是否存在
	//isExist, err := robotgo.PidExists(pIdInfo.PID)
	//if err != nil {
	//	return false, err
	//}
	//if !isExist {
	//	//进程不存在
	//	return false, nil
	//}
	//进程存在，则判断进程名称是否相同
	//fmt.Println("pid exists is", isExist)
	//name, err := robotgo.FindName(pIdInfo.PID)
	//if err != nil {
	//	if err.Error() == "process does not exist" {
	//		return false, nil
	//	}
	//	//fmt.Println("未找到进程", err.Error())
	//	return false, err
	//}
	//fmt.Println("进程名称1", name)

	//localName, err := robotgo.FindName(syscall.Getpid())
	//if err != nil {
	//	return false, err
	//}
	//fmt.Println("进程名称2", localName)
	//if name != localName {
	//	return false, nil
	//}
	//对比可执行文件绝对地址
	if os.Args[0] != pIdInfo.Abs {
		return false, nil
	}
	//
	//err = robotgo.ActivePid(pIdInfo.PID)
	//if err != nil {
	//	return false, err
	//}
	return true, nil
}

/*
程序路径及进程id灯信息
*/
type ProcessIdInfo struct {
	Abs  string //程序绝对路径
	Name string //进程名称
	PID  int    //进程id编号
}
