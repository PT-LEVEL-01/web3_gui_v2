package restart

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"web3_gui/chain/config"
	"web3_gui/libp2parea/adapter/engine"
)

const (
	//期望的重启参数
	envRestartArgKey = "envRestartArgKey"
)

var mainPID = int(0)

// 通知重启程序
// arg 期望重启参数: load 或 空
func NotifyRestart(arg string) error {
	//获取主进程信息
	execName, err := os.Executable()
	if err != nil {
		return err
	}
	//获取主进程PID
	mainPID = os.Getpid()

	args := []string{execName, config.RestartCommand}
	restartArgEnv := fmt.Sprintf("%s=%s", envRestartArgKey, arg)
	_, err = forkProcess(execName, args, restartArgEnv)
	if err != nil {
		return err
	}

	//退出并释放主进程
	if err := exitProcess(); err != nil {
		return err
	}

	return nil
}

// 启动程序
func Restart() error {
	//获取主进程信息
	execName, err := os.Executable()
	if err != nil {
		return err
	}
	//获取主进程PID
	mainPID = os.Getpid()

	//重启参数
	arg := os.Getenv(envRestartArgKey)
	arg = strings.TrimSpace(arg)
	args := []string{execName, arg}

	_, err = forkProcess(execName, args, "")
	if err != nil {
		return err
	}

	//退出并释放主进程
	if err := exitProcess(); err != nil {
		return err
	}

	return nil
}

// 退出进程
func exitProcess() error {
	engine.Log.Info("Exit Process PID:%d", mainPID)
	if proc, err := os.FindProcess(mainPID); err == nil {
		proc.Kill()
		//释放进程资源
		proc.Release()
	} else {
		return err
	}

	//等待一下程序彻底释放
	time.Sleep(time.Second * 5)

	//再次检查端口
	conn, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", config.WebPort))
	if err != nil {
		return err
	}

	conn.Close()
	return nil
}

// 启动子进程
func forkProcess(execName string, args []string, restartArgEnv string) (*os.Process, error) {
	envs := os.Environ()
	if restartArgEnv != "" {
		envs = append(envs, restartArgEnv)
	}

	p, err := os.StartProcess(execName, args, &os.ProcAttr{
		Dir: filepath.Dir(execName),
		Env: envs,
		Files: []*os.File{
			os.Stdin,
			os.Stdout,
			os.Stderr,
		},
		Sys: &syscall.SysProcAttr{},
	})
	if err != nil {
		return nil, err
	}

	engine.Log.Info("Process PID %d Starting Child Process PID %d With Args:%v", mainPID, p.Pid, args)

	return p, err
}
