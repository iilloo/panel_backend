package services

import (
	"bufio"
	"encoding/json"
	"panel_backend/global"
	"panel_backend/models"

	"github.com/gorilla/websocket"
	"github.com/google/uuid"
)

func HandleOrder(order string, conn *websocket.Conn) {
	//如果没有启动bash进程，要启动bash进程
	if global.Bash.CMD.Process == nil{
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
	count, err:= global.Bash.StdinPipe.Write([]byte(fullOrder))
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
		if line == "END_OF_COMMAND_" + global.Bash.UUID + "\n" {
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
