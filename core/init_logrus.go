package core

import (
	"io"
	_ "panel_backend/global"
	"panel_backend/utils/customLog/customHook"

	"github.com/sirupsen/logrus"
)

func InitLogger(logPath string, appName string) *logrus.Logger{
	log := logrus.New()
	//debug以上的都输出
	log.SetLevel(5)
	//设置不在终端输出，自定义钩子函数中再设置输出到终端
	log.SetOutput(io.Discard)
	//先拿到原始的msg对满足条件的msg进行格式化
	customhook.InitFormat(log)
	//把格式化后的日志写入文件
	customhook.InitFile(log, logPath, appName)
	return log
}