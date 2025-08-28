package web

import (
	"github.com/astaxie/beego"
	"os/exec"
	"runtime"
	"strconv"
	"web3_gui/libp2parea/v1/sdk/config"
	routers "web3_gui/libp2parea/v1/sdk/web/routers"
)

// Start Start
func Start(libp2pareaConfig string) {
	// 解析外部应用传递的参数
	config.ParseConfig(libp2pareaConfig)

	beego.BConfig.WebConfig.Session.SessionOn = false
	beego.BConfig.Listen.HTTPPort = int(config.WebPort)
	beego.BConfig.Listen.HTTPSAddr = config.WebAddr
	beego.BConfig.WebConfig.Session.SessionName = "libp2parea"
	beego.BConfig.WebConfig.Session.SessionGCMaxLifetime = 3600
	beego.BConfig.CopyRequestBody = true
	beego.BConfig.WebConfig.TemplateLeft = "<%"
	beego.BConfig.WebConfig.TemplateRight = "%>"
	// beego.PprofOn = true

	//home
	//	beego.SetStaticPath("/static", `D:\workspaces\go\src\github.com/prestonTao/abcchainnewaccount\web\static`)
	//	beego.BConfig.WebConfig.ViewsPath = `D:\workspaces\go\src\github.com/prestonTao/abcchainnewaccount\web\views`

	//inc
	//	beego.SetStaticPath("/static", `D:\workspace\src\github.com/prestonTao/abcchainnewaccount\web\static`)
	//	beego.BConfig.WebConfig.ViewsPath = `D:\workspace\src\github.com/prestonTao/abcchainnewaccount\web\views`
	// beego.SetStaticPath("/static", config.Web_path_static)
	beego.BConfig.WebConfig.ViewsPath = config.Web_path_views
	beego.SetStaticPath("/static", config.Web_path_static)
	routers.Start()
	//	go openLocalWeb()
	// beego.Run()
}

// Open calls the OS default program for uri
func openLocalWeb() {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start http://127.0.0.1:"+strconv.Itoa(int(config.WebPort))+"/")
	case "darwin":
		cmd = exec.Command("open", "http://127.0.0.1:"+strconv.Itoa(int(config.WebPort))+"/")
	case "linux":
		cmd = exec.Command("xdg-open", "http://127.0.0.1:"+strconv.Itoa(int(config.WebPort))+"/")

	}
	err := cmd.Start()
	if err != nil {
		// fmt.Printf("启动页面的时候发生错误:%s", err.Error())
	}
}
