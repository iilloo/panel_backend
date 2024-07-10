package services

import (
	_"bufio"
	_"encoding/json"
	"fmt"

	"panel_backend/global"
	_"panel_backend/models"

	"github.com/creack/pty"
	_"github.com/google/uuid"
	"github.com/gorilla/websocket"
)
func HandleOrder(order string, conn *websocket.Conn) {
	//如果没有启动bash进程，启动bash进程
	//适用于第一次启动bash进程
	if global.Bash.CMD.Process == nil {
		fmt.Println("bash进程未启动")
		ptmx, err := pty.Start(global.Bash.CMD)
		global.Bash.Ptmx = ptmx
		if err != nil {
			global.Log.Errorf("启动bash进程以及伪终端pty失败: %s", err.Error())
			return
		}
		pty.Setsize(global.Bash.Ptmx, &pty.Winsize{
			Cols: global.Bash.Cols,
			Rows: global.Bash.Rows,
		})
		
		//继续检测可能的输出
		go func() {
			buf := make([]byte, 1024)
			for {
				n, err := global.Bash.Ptmx.Read(buf)
				if err != nil {
					global.Log.Errorf("处理交互性命令读取伪终端pty输出失败: %s", err.Error())
					return
				}
				// fmt.Printf("cmdStdout: %s\n", string(buf[:n]))
				response := messageString("cmdStdout", string(buf[:n]))
				global.Log.Infof("cmdStdout: %s\n", response)
				conn.WriteMessage(websocket.TextMessage, response)
				// select {
				// case <-global.Bash.StopInPutChan:
				// 	global.Log.Infof("关闭持续输出协程\n")
				// 	return
				// default:
				// 	n, err := global.Bash.Ptmx.Read(buf)
				// 	if err != nil {
				// 		global.Log.Errorf("处理交互性命令读取伪终端pty输出失败: %s", err.Error())
				// 		return
				// 	}
				// 	// fmt.Printf("cmdStdout: %s\n", string(buf[:n]))
				// 	response := messageString("cmdStdout", string(buf[:n]))
				// 	conn.WriteMessage(websocket.TextMessage, response)
				// }
			}
		}()
	}
	//用于判断bash进程关闭过后是否重新启动bash进程
	// global.Log.Infof("global.Bash.CMD.Process != nil\n")
	// if global.Bash.CMD.Process != nil && global.Bash.CMD.Process.Signal(syscall.Signal(0)) != nil {
	// 	global.Log.Infof("bash进程已经关闭，重新启动bash进程\n")
	// 	global.Bash.CMD = exec.Command("bash")
	// 	ptmx, err := pty.Start(global.Bash.CMD)
	// 	global.Bash.Ptmx = ptmx
	// 	if err != nil {
	// 		global.Log.Errorf("重新启动bash进程以及伪终端pty失败: %s", err.Error())
	// 		return
	// 	}
	// 	//继续检测可能的输出
	// 	go func() {
	// 		buf := make([]byte, 1024)
	// 		for {
	// 			n, err := global.Bash.Ptmx.Read(buf)
	// 				if err != nil {
	// 					global.Log.Errorf("处理交互性命令读取伪终端pty输出失败: %s", err.Error())
	// 					return
	// 				}
	// 				// fmt.Printf("cmdStdout: %s\n", string(buf[:n]))
	// 				response := messageString("cmdStdout", string(buf[:n]))
	// 				conn.WriteMessage(websocket.TextMessage, response)
	// 		}
	// 	}()
	// }
	// if global.Bash.Cols != global.Bash.ColsPre && global.Bash.Rows != global.Bash.RowsPre {
	// 	pty.Setsize(global.Bash.Ptmx, &pty.Winsize{
	// 		Cols: global.Bash.Cols,
	// 		Rows: global.Bash.Rows,
	// 	})
	// 	global.Bash.ColsPre = global.Bash.Cols
	// 	global.Bash.RowsPre = global.Bash.Rows
	// 	global.Log.Infof("pty设置大小成功Cols:%d,Rows:%d\n", global.Bash.Cols, global.Bash.Rows)
	// }
	global.Log.Infof("order: %s\n", order)
	_, err := global.Bash.Ptmx.Write([]byte(order))
	if err != nil {
		global.Log.Errorf("写入伪终端pty失败: %s", err.Error())
		return
	}

}
