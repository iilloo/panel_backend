package core

import (
	"os/exec"
	"panel_backend/global"
	"panel_backend/models"
)

func InitCMD() *models.Bash {
	cmd := exec.Command("bash")
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		global.Log.Errorf("获取bash进程输入管道失败: %s", err.Error())
		return nil
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		global.Log.Errorf("获取bash进程输出管道失败: %s", err.Error())
		return nil
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		global.Log.Errorf("获取bash进程错误输出管道失败: %s", err.Error())
		return nil
	}
	return &models.Bash{
		CMD: cmd,
		StdinPipe: stdinPipe,
		StdoutPipe: stdoutPipe,
		StderrPipe: stderrPipe,
		UUID: "",
		Order: "",
	}
}