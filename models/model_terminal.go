package models

import (
	_"io"
	"os"
	"os/exec"
	"os/user"
)

type Bash struct {
	CMD *exec.Cmd
	Usr *user.User
	Order string
	Ptmx *os.File
	StopInPutChan chan bool
	Cols uint16
	Rows uint16

}
type PtyInfo struct {
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`

}