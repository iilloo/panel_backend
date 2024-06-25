package models

import (
	"io"
	"os"
	"os/exec"
)

type Bash struct {
	CMD *exec.Cmd
	StdinPipe io.WriteCloser
	StdoutPipe io.ReadCloser
	StderrPipe io.ReadCloser
	UUID string
	Order string
	Ptmx *os.File
	StopInPutChan chan bool
}
type PtyInfo struct {
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`

}