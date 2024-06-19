package customformat

import (
	"bytes"
	"fmt"

	"github.com/sirupsen/logrus"
)

type BlankFormatter struct {

}

//实现Format接口
func (f *BlankFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer == nil {
		b = &bytes.Buffer{}
	} else {
		b = entry.Buffer
	}
	fmt.Fprintf(b, entry.Message)
	return b.Bytes(), nil
}