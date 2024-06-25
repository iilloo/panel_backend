package tests

import (
	"os/exec"
	"syscall"
	"testing"

	"github.com/creack/pty"
)

func TestTermianl(t *testing.T) {
	cmd := exec.Command("bash")
	ptymx, _ := pty.Start(cmd)
	ptymx.Close()
	cmd.Process.Signal(syscall.SIGTERM)
	cmd.Wait()

	if cmd.Process != nil {
		t.Log("cmd.Process is !nil")
	}
	if cmd.Process.Signal(syscall.Signal(0)) != nil {
		t.Log("bash process has been closed")
	}
	cmd = exec.Command("bash")
	ptymx, err := pty.Start(cmd)
	if err != nil {
		t.Error("pty.Start error")
	}
	_, err = ptymx.Write([]byte("ls\n"))
	if err != nil {
		t.Error("ptymx.Write error")
	}
	ptymx.Close()

}