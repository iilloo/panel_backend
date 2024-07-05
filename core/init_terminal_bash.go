package core

import (
	"os/exec"
	"panel_backend/global"
	"panel_backend/models"
)

func InitCMD() *models.Bash {
	cmd := exec.Command("bash")
	// stdinPipe, err := cmd.StdinPipe()
	// if err != nil {
	// 	global.Log.Errorf("获取bash进程输入管道失败: %s", err.Error())
	// 	return nil
	// }
	// stdoutPipe, err := cmd.StdoutPipe()
	// if err != nil {
	// 	global.Log.Errorf("获取bash进程输出管道失败: %s", err.Error())
	// 	return nil
	// }
	// stderrPipe, err := cmd.StderrPipe()
	// if err != nil {
	// 	global.Log.Errorf("获取bash进程错误输出管道失败: %s", err.Error())
	// 	return nil
	// }
	stopChan := make(chan bool, 1)
	global.Log.Infof("bash进程初始化成功\n")
	return &models.Bash{
		CMD:        cmd,
		StdinPipe:  nil,
		StdoutPipe: nil,
		StderrPipe: nil,
		UUID:       "",
		Order:      "",
		Ptmx:       nil,
		StopInPutChan: stopChan,
		Cols:       160,
		Rows:       42,
	}
}
