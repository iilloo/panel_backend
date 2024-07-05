package services

import (
	"bufio"
	"encoding/json"
	"fmt"

	"panel_backend/global"
	"panel_backend/models"

	"github.com/creack/pty"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func HandleOrder(order string, conn *websocket.Conn) {
	//如果没有启动bash进程，要启动bash进程
	if global.Bash.CMD.Process == nil {
		err := global.Bash.CMD.Start()
		global.Log.Infof("bash进程启动\n")
		if err != nil {
			global.Log.Errorf("启动bash进程失败: %s", err.Error())
			return
		}
	}
	global.Log.Infof("order: %s\n", order)
	global.Bash.Order = order
	// 生成一个唯一的ID
	global.Bash.UUID = uuid.New().String()
	fullOrder := order + "; echo END_OF_COMMAND_" + global.Bash.UUID + "\n"
	//将包含命令和唯一ID的命令发送到bash进程的标准输入
	count, err := global.Bash.StdinPipe.Write([]byte(fullOrder))
	if err != nil {
		global.Log.Errorf("写入bash进程标准输入失败: %s", err.Error())
		return
	}
	global.Log.Infof("count: %d\n", count)
	reader := bufio.NewReader(global.Bash.StdoutPipe)
	for {
		global.Log.Infof("aaaaaaaaaaaaaaaaaaa\n")
		line, err := reader.ReadString('\n')
		global.Log.Infof("bbbbbbbbbbbbbbbbb\n")
		if err != nil {
			global.Log.Errorf("读取bash进程标准输出失败: %s", err.Error())
			return
		}
		if line == "END_OF_COMMAND_"+global.Bash.UUID+"\n" {
			response := models.Message{Type: "cmdStdout", Data: "END_OF_COMMAND_" + global.Bash.Order}
			jsonMsg, err := json.Marshal(response)
			if err != nil {
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, jsonMsg); err != nil {
				return
			}
			break
		}
		msg := models.Message{Type: "cmdStdout", Data: line}
		jsonMsg, err := json.Marshal(msg)
		if err != nil {

			return
		}
		if err := conn.WriteMessage(websocket.TextMessage, jsonMsg); err != nil {

			return
		}
	}

}

func HandleOrder_1(order string, conn *websocket.Conn) {
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
