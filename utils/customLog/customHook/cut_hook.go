package customhook

import (
	"fmt"
	"os"
	_ "panel_backend/global"
	"time"

	"github.com/sirupsen/logrus"
)

// 自定义hook,实现logrus日志按日期和类型分割成多个文件
type LogDataKindCutHook struct {
	file      *os.File
	errorFile *os.File
	warnFile  *os.File
	debugFile *os.File
	infoFile  *os.File
	logPath   string
	fileData  string
	appName   string
}

// 实现Levels接口
func (hook *LogDataKindCutHook) Levels() []logrus.Level {
	//作用于全部等级
	return logrus.AllLevels

}

// 根据已有的日志等级进一步根据时间进行日志分割，并写入相关文件
func (hook *LogDataKindCutHook) timeCut(file *os.File, time string, logStr string, postfix string) {
	//时间
	// if postfix != "debug" {
	// 	os.Stdout.WriteString(logStr)
	// }
	if time == hook.fileData {
		file.Write([]byte(logStr))
	} else {
		hook.fileData = time
		file.Close()
		os.MkdirAll(fmt.Sprintf("%s/%s", hook.logPath, time), os.ModePerm)
		filename := fmt.Sprintf("%s/%s/%s_%s.log", hook.logPath, time, hook.appName, postfix)
		file, _ = os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		file.Write([]byte(logStr))
	}
}
func  outPutToTerminal(logStr string) {
	os.Stdout.WriteString(logStr)
}

// Fire(entry *logrus.Entry) error接口
func (hook *LogDataKindCutHook) Fire(entry *logrus.Entry) error {
	time := entry.Time.Format("2006-01-02")
	logStr, err := entry.String()
	if err != nil {
		return err
	}
	//对日志的等级进行分割
	logLevel := entry.Level
	switch logLevel {
	case logrus.ErrorLevel:
		hook.timeCut(hook.errorFile, time, logStr, "err")
		outPutToTerminal(logStr)
	case logrus.WarnLevel:
		hook.timeCut(hook.warnFile, time, logStr, "warn")
		outPutToTerminal(logStr)
	case logrus.DebugLevel:
		hook.timeCut(hook.debugFile, time, logStr, "debug")
	case logrus.InfoLevel:
		hook.timeCut(hook.infoFile, time, logStr, "info")
		outPutToTerminal(logStr)
	}
	hook.timeCut(hook.file, time, logStr, "all")
	return nil
}

// 初始化函数
func InitFile(log *logrus.Logger, logPath, appName string) {
	//获取当前时间
	fileData := time.Now().Format("2006-01-02")
	//创建目录
	err := os.MkdirAll(fmt.Sprintf("%s/%s", logPath, fileData), os.ModePerm)
	if err != nil {
		logrus.Error(err)
		return
	}
	//初始化创建各个日志文件
	filename := fmt.Sprintf("%s/%s/%s", logPath, fileData, appName)
	file, _ := os.OpenFile(filename+"_all.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	errorFile, _ := os.OpenFile(filename+"_err.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	warnFile, _ := os.OpenFile(filename+"_warn.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	debugFile, _ := os.OpenFile(filename+"_debug.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	infoFile, _ := os.OpenFile(filename+"_info.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)

	filehook := LogDataKindCutHook{file,
		errorFile, warnFile, debugFile, infoFile,
		logPath, fileData, appName}
	//添加钩子函数
	log.AddHook(&filehook)
}
