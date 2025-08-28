package tray

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	"sync"
	"time"
	"web3_gui/utils"
)

var trayLock *sync.RWMutex = new(sync.RWMutex) //
var app *application.App                       //
var traySwitch = false                         //托盘开关
var msgChan = make(chan bool, 1000)            //消息管道
var once = new(sync.Once)                      //
var sTray *application.SystemTray              //托盘
var msgType = make(map[string]int)             //当注册的消息key都删除完后，托盘图标才会停止闪烁。key:string=注册消息key;value:any=;

//var systemTrayPngData []byte
//var systemTrayLightPngData []byte

/*
初始化系统托盘
*/
func InitTray(a *application.App) {
	trayLock.Lock()
	app = a
	trayLock.Unlock()
	once.Do(flicker)
}

/*
显示系统托盘
*/
func OpenSystemTray() {
	if app == nil {
		utils.Log.Error().Msgf("app变量未初始化")
		return
	}
	trayLock.Lock()
	if traySwitch {
		trayLock.Unlock()
		return
	}
	traySwitch = true
	//系统托盘
	systemTray := app.SystemTray.New()
	systemTray.SetIcon(systemTrayPngData)
	myMenu := app.NewMenu()
	myMenu.Add("显示").OnClick(func(_ *application.Context) {
		ShowWindow()
	})
	myMenu.Add("隐藏").OnClick(func(_ *application.Context) {
		app.Hide()
	})
	myMenu.Add("退出").OnClick(func(_ *application.Context) {
		app.Quit()
	})
	systemTray.SetMenu(myMenu)
	//systemTray.AttachWindow(window).WindowOffset(0)

	systemTray.OnClick(func() {
		utils.Log.Info().Str("鼠标", "OnClick").Send()
		window := app.Window.Current() //.CurrentWindow()
		if window.IsVisible() {
			app.Hide()
			return
		}
		ShowWindow()
	})
	//systemTray.OnMouseEnter(func() {
	//	utils.Log.Info().Str("鼠标", "OnMouseEnter").Send()
	//	app.Show()
	//})

	sTray = systemTray
	trayLock.Unlock()
}

func ShowWindow() {
	window := app.Window.Current() //.CurrentWindow()
	//如果是最小化，则取消最小化
	if window.IsMinimised() {
		window.UnMinimise()
		return
	}
	//如果是隐藏，则设置为显示
	if !window.IsVisible() {
		window.Show()
		return
	}
	//如果是显示状态，则设置为最顶层
	window.SetAlwaysOnTop(true)
	time.Sleep(time.Second / 5)
	window.SetAlwaysOnTop(false)
}

/*
删除并销毁系统托盘
*/
func CloseSystemTray() {
	trayLock.Lock()
	if !traySwitch {
		trayLock.Unlock()
		return
	}
	traySwitch = false
	//销毁系统托盘
	sTray.Destroy()
	trayLock.Unlock()
	//utils.Log.Info().Msgf("关闭托盘")
}

/*
系统托盘闪烁协程
*/
func flicker() {
	flickerV2()
	return
	go func() {
		for {
			have := <-msgChan
			if !have {
				continue
			}
		Prime: //循环标签
			for {
				select {
				case have := <-msgChan:
					if !have {
						break Prime
					}
				default:
				}
				time.Sleep(time.Second / 3)
				sTray.SetIcon(systemTrayNewMsgPngData)
				time.Sleep(time.Second / 3)
				sTray.SetIcon(systemTrayPngData)
			}
		}
	}()
}

/*
系统托盘闪烁协程
这版仅图标小红点提醒,托盘不闪烁
*/
func flickerV2() {
	go func() {
		for {
			have := <-msgChan
			if have {
				sTray.SetIcon(systemTrayNewMsgPngData)
			} else {
				sTray.SetIcon(systemTrayPngData)
			}
		}
	}()
}

//
//type TrayLifeCycle struct {
//	closeChan chan bool
//}
//
//func NewTrayLifeCycle() *TrayLifeCycle {
//	return &TrayLifeCycle{
//		closeChan: make(chan bool, 1),
//	}
//}
//
//func (this *TrayLifeCycle) onReady() {
//	systray.SetIcon(Data)
//	systray.SetTitle(config.GUITitle)
//	systray.SetTooltip(config.GUITitle)
//	mShow := systray.AddMenuItem("显示", "显示窗口")
//	mHide := systray.AddMenuItem("隐藏", "隐藏窗口")
//	systray.AddSeparator()
//	mQuit := systray.AddMenuItem("退出", "退出程序")
//
//	//kernel32 := syscall.NewLazyDLL("kernel32.dll")
//	//user32 := syscall.NewLazyDLL("user32.dll")
//	//// https://docs.microsoft.com/en-us/windows/console/getconsolewindow
//	//getConsoleWindows := kernel32.NewProc("GetConsoleWindow")
//	//// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-showwindowasync
//	//showWindowAsync := user32.NewProc("ShowWindowAsync")
//	//consoleHandle, r2, err := getConsoleWindows.Call()
//	//if consoleHandle == 0 {
//	//	fmt.Println("Error call GetConsoleWindow: ", consoleHandle, r2, err)
//	//}
//
//	go this.listen(mShow, mHide, mQuit)
//
//}
//
//func (this *TrayLifeCycle) listen(mShow, mHide, mQuit *systray.MenuItem) {
//	for {
//		select {
//		case <-mShow.ClickedCh:
//			//utils.Log.Info().Msgf("显示")
//			//mShow.Disable()
//			//mHide.Enable()
//			runtime.WindowShow(appCtx)
//		//r1, r2, err := showWindowAsync.Call(consoleHandle, 5)
//		//if r1 != 1 {
//		//	fmt.Println("Error call ShowWindow @SW_SHOW: ", r1, r2, err)
//		//}
//		case <-mHide.ClickedCh:
//			//	utils.Log.Info().Msgf("隐藏")
//			//	mHide.Disable()
//			//	mShow.Enable()
//			runtime.WindowMinimise(appCtx)
//		//r1, r2, err := showWindowAsync.Call(consoleHandle, 0)
//		//if r1 != 1 {
//		//	fmt.Println("Error call ShowWindow @SW_HIDE: ", r1, r2, err)
//		//}
//		case <-mQuit.ClickedCh:
//			//utils.Log.Info().Msgf("退出")
//			systray.Quit()
//			runtime.Quit(appCtx)
//		case <-this.closeChan:
//			//utils.Log.Info().Msgf("结束")
//			return
//		}
//	}
//}
//
//func (this *TrayLifeCycle) onExit() {
//	// clean up here
//	//utils.Log.Info().Msgf("清理 start")
//	this.closeChan <- false
//	//utils.Log.Info().Msgf("清理 end")
//}

/*
发送通知
通过通知控制托盘是否闪烁
@flicker    bool      是否闪烁；true=闪烁；false=取消闪烁；
@key        string    注册和取消的key，当所有的key都取消完后，才会停止闪烁
*/
func SendNotice(flicker bool, key string) {
	//utils.Log.Info().Msgf("发送闪烁通知:%t %s", flicker, key)
	if key == "" {
		return
	}
	//先判断开关是否打开
	open := false
	trayLock.RLock()
	defer trayLock.RUnlock()
	open = traySwitch
	if !open {
		return
	}
	//再判断界面是否激活状态
	//normal := runtime.WindowIsNormal(appCtx)
	//utils.Log.Info().Msgf("界面状态:%t", normal)
	//x, y := runtime.WindowGetPosition(appCtx)
	//utils.Log.Info().Msgf("坐标:%d %d", x, y)
	if flicker {
		msgType[key] = 0
	} else {
		delete(msgType, key)
		if len(msgType) != 0 {
			return
		}
	}

	//尝试往里面放
	select {
	case msgChan <- flicker:
		return
	default:
		//放不进去就取出来再放
		select {
		case <-msgChan:
			//待取出后再次放入
			select {
			case msgChan <- flicker:
			default:
			}
		default:
		}
	}
}

/*
关闭回调函数
当有托盘时,主进程窗口只能隐藏，不能关闭
*/
func MastertWindowCloseCallback(window *application.WebviewWindow) bool {
	trayLock.Lock()
	if traySwitch {
		//utils.Log.Info().Msgf("打开托盘")
		//托盘打开，不能退出
		trayLock.Unlock()
		window.Hide()
		return false
	}
	//托盘未打开，可以退出
	return true
}
