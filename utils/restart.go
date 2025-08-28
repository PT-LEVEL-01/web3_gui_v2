package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

const (
	//期望的重启参数
	envRestartArgKey = "envRestartArgKey" //
	RestartCommand   = "restart"          //重启参数
)

// 通知重启程序
// arg 期望重启参数: load 或 空
func NotifyRestart() error {
	//params := strings.Join(os.Args, " ")
	//fmt.Println("启动命令", params)
	arg := strings.Join(os.Args[1:], " ")
	//获取主进程信息
	execName, err := os.Executable()
	if err != nil {
		return err
	}
	//获取主进程PID
	mainPid := os.Getpid()

	args := []string{execName, RestartCommand}
	restartArgEnv := fmt.Sprintf("%s=%s", envRestartArgKey, arg)
	_, err = forkProcess(execName, args, restartArgEnv)
	if err != nil {
		return err
	}

	//退出并释放主进程
	if err := exitProcess(mainPid); err != nil {
		return err
	}
	return nil
}

// 退出进程
func exitProcess(pid int) error {
	fmt.Println("退出进程ID", pid)
	//utils.Log.Info().Msgf("Exit Process PID:%d", mainPID)
	if proc, err := os.FindProcess(pid); err == nil {
		proc.Kill()
		//释放进程资源
		proc.Release()
	} else {
		return err
	}
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
	//fmt.Println("进程id", mainPID, "子进程id", p.Pid, "参数", args)
	//utils.Log.Info().Msgf("Process PID %d Starting Child Process PID %d With Args:%v", mainPID, p.Pid, args)

	return p, err
}
