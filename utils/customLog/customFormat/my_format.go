package customformat

import (
	"bytes"
	"fmt"

	"github.com/sirupsen/logrus"
)

//根据日志等级得到对应的颜色头和颜色尾
func  setColorLevel (level logrus.Level) (string, string) {
	tag := ""
	switch level {
	case logrus.ErrorLevel :
		tag = "1"
	case logrus.WarnLevel :
		tag = "3"
	case logrus.InfoLevel :
		tag = "4"
	case logrus.DebugLevel :
		tag = "2"
	}
	return "\033[4" + tag + "m", "\033[0m"
}

//自定义日志格式,只作用于error和warn
type MyFormatter struct {
}


func (f *MyFormatter) Format (entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer == nil {
		b = &bytes.Buffer{}
	} else {
		b = entry.Buffer
	}
	//设置level的对应的颜色
	startColor, endColor := setColorLevel(entry.Level)
	//时间格式化
	time := entry.Time.Format("2006-01-02 15:04:06")
	//log信息的文件路径和行号(位置)
	fileVal := fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line)
	fmt.Fprintf(b, "   |%s %s %s| %s Msg:[%s] Path: %s\n",startColor, entry.Level, endColor, time, entry.Message, fileVal)
	return b.Bytes(), nil
}