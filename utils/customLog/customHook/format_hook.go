package customhook

import (
	_ "bytes"

	_ "fmt"
	_"panel_backend/global"
	"panel_backend/utils/customLog/customFormat"
	_ "path/filepath"

	"github.com/sirupsen/logrus"
)

//自定义格式化的钩子结构
type MyFormatHook struct {
	formatter logrus.Formatter
    levels    []logrus.Level
}

// 重写Fire方法对原始日志信息进行加工
func (h *MyFormatHook) Fire(entry *logrus.Entry) error {
	for _, v := range h.levels {
		if v == entry.Level {
			formatMsg, _ := h.formatter.Format(entry)
			entry.Message = string(formatMsg)
			break
		}
	}
    return nil
}
// 重写Levels方法确定该钩子作用的日志等级
func (h *MyFormatHook) Levels() []logrus.Level {
    return h.levels
}

func InitFormat(log *logrus.Logger) {
	// 让其可以使用某些方法
	log.SetReportCaller(true)
	//用空的自定义格式化结构提纯出真正的日志消息，不让其携带除日志消息以外的内容
	log.SetFormatter(&customformat.BlankFormatter{})
	// 得到一个自定义的format
	formatter := &customformat.MyFormatter{}
	hook := &MyFormatHook{
        formatter: formatter,
        levels:    []logrus.Level{logrus.ErrorLevel, logrus.WarnLevel},
    }
	log.AddHook(hook)
}