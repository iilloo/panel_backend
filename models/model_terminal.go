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
	Cols uint16
	Rows uint16
	ColsPre uint16
	RowsPre uint16
}
type PtyInfo struct {
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`

}