package core

import (
	"os/exec"
	"os/user"
	"panel_backend/global"
	"panel_backend/models"
)

func InitCMD() *models.Bash {
	cmd := exec.Command("bash")
	usr, err := user.Current()
	if err != nil {
		global.Log.Errorf("获取当前用户失败: %s", err.Error())
		return nil
	}
	cmd.Dir = usr.HomeDir // 设置工作目录为用户主目录
	
	stopChan := make(chan bool, 1)
	global.Log.Infof("bash进程初始化成功\n")
	return &models.Bash{
		CMD:           cmd,
		Usr:           usr,
		Order:         "",
		Ptmx:          nil,
		StopInPutChan: stopChan,
		Cols:          160,
		Rows:          42,
	}
}
