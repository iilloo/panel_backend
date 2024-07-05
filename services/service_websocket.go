package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"panel_backend/global"
	"panel_backend/utils/jwts"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"panel_backend/models"
)

// 升级http协议为websocket协议
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		//允许所有的跨域请求
		return true
	},
}

func verifyId(msgString []byte) (bool, string) {
	var msg models.Message
	//解析json数据
	if err := json.Unmarshal(msgString, &msg); err != nil {
		// global.Log.Errorf("websocket解析身份认证信息失败:[%s]\n", err.Error())
		return false, ""
	}
	if msg.Type != "token" {
		// global.Log.Errorf("websocket身份认证信息类型错误\n")
		return false, ""

	}
	//在这里进行身份认证
	if msg.Data == "" {
		return false, ""
	}
	_, newTokenString, err := jwts.ParseToken(msg.Data.(string))
	if err != nil {
		// global.Log.Errorf("websocket身份认证失败:[%s]\n", err.Error())
		return false, ""
	}
	//设置刷新后的token
	if newTokenString != msg.Data {

		return true, newTokenString
	}
	return true, ""
}
func messageString(msgType string, data interface{}) []byte {
	var response models.Message
	response.Type = msgType
	response.Data = data
	responseByte, _ := json.Marshal(response)
	return responseByte
}
func MessageString(msgType string, data interface{}) []byte {
	var response models.Message
	response.Type = msgType
	response.Data = data
	responseByte, _ := json.Marshal(response)
	return responseByte
}
func WS(c *gin.Context) {
	//websocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "websocket连接失败",
		})
		global.Log.Errorf("websocket连接失败:[%s]\n", err.Error())
		return
	}
	global.Log.Infof("websocket连接成功\n")
	//返回给前端的消息
	response := messageString("success", "websocket连接成功")
	conn.WriteMessage(websocket.TextMessage, response)

	defer conn.Close()

	//建立连接后，读取前端传过来的身份认证信息10次机会
	//设置身份验证消息的超时时间为5秒
	conn.SetReadDeadline(time.Now().Add(5 * time.Second)) // 5秒超时

	for i := 0; i < 10; i++ {
		//如果前端迟迟不来消息则会阻塞在这里
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			global.Log.Errorf("websocket读取MESSAGE失败:[%s]\n", err.Error())
			//给前端回复错误信息
			response := messageString("error", "读取MESSAGE失败")
			conn.WriteMessage(messageType, response)
			if i == 9 {
				//如果读取信息失败则关闭连接
				conn.Close()
				return
			}
		}
		isSuccess, newTokenString := verifyId(msg)
		if !isSuccess {
			global.Log.Errorf("websocket_TOKEN身份认证失败\n")
			//给前端回复错误信息
			response := messageString("error", "websocket身份认证失败")
			conn.WriteMessage(messageType, response)
			if i == 9 {
				//如果认证失败则关闭连接
				conn.Close()
				return
			}
		} else {
			//给前端回复成功信息
			response := messageString("success", "websocket身份认证成功")
			global.Log.Infof("websocket身份认证成功\n")
			conn.WriteMessage(messageType, response)
			//如果有新的token则返回给前端
			if newTokenString != "" {
				newTokenByte := messageString("token", newTokenString)
				conn.WriteMessage(messageType, newTokenByte)
			}
			//取消超时时间
			conn.SetReadDeadline(time.Time{})
			//如果认证成功则退出循环
			break
		}

	}

	//通过身份验证之后，开始正常的消息处理
	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			global.Log.Errorf("websocket读取信息失败:[%s]\n", err.Error())
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				global.Log.Errorf("websocket连接异常关闭\n")

				//关闭pty伪终端
				global.Bash.Ptmx.Close()
				global.Bash.Ptmx = nil
				global.Log.Infof("Ptmx伪终端关闭\n")
				//重新启动bash进程
				global.Bash.CMD = exec.Command("bash")
				global.Log.Infof("重新启动bash进程\n")
				// //关闭bash进程
				// if err := global.Bash.CMD.Process.Signal(syscall.SIGTERM); err != nil {
				// 	global.Log.Errorf("无法发送 SIGTERM 信号: %v\n", err)
				// 	err := global.Bash.CMD.Process.Signal(syscall.SIGKILL)
				// 	if err != nil {
				// 		global.Log.Errorf("无法发送 SIGKILL 信号: %v\n", err)
				// 	}

				// }
				// global.Bash.CMD.Wait()
				// global.Log.Infof("bash进程关闭\n")

				return
			} else if websocket.IsCloseError(err, websocket.CloseGoingAway) {
				global.Log.Errorf("websocket连接被前端正常关闭\n")
				//关闭pty伪终端
				global.Bash.Ptmx.Close()
				global.Bash.Ptmx = nil
				global.Log.Infof("Ptmx伪终端关闭\n")
				//重新启动bash进程
				global.Bash.CMD = exec.Command("bash")
				global.Log.Infof("重新启动bash进程\n")
				// //关闭bash进程
				// if err := global.Bash.CMD.Process.Signal(syscall.SIGKILL); err != nil {
				// 	global.Log.Errorf("无法发送 SIGTERM 信号: %v\n", err)
				// 	err := global.Bash.CMD.Process.Signal(syscall.SIGKILL)
				// 	if err != nil {
				// 		global.Log.Errorf("无法发送 SIGKILL 信号: %v\n", err)
				// 	}
				// }
				// global.Log.Infof("aaaaaaaaaaaaaaaaaaa\n")
				// global.Bash.CMD.Wait()
				// global.Log.Infof("bbbbbbbbbbbbbbbbbbb\n")
				// global.Log.Infof("bash进程关闭\n")
				return
			}
			//给前端回复错误信息
			// conn.WriteMessage(messageType, []byte("读取信息失败"))
			response := messageString("error", "读取信息失败")
			conn.WriteMessage(messageType, response)
			continue
		}
		// 在这里处理来自前端的数据
		global.Log.Debugf("收到消息: %s\n", msg)
		var message models.Message
		if err := json.Unmarshal(msg, &message); err != nil {
			global.Log.Errorf("websocket解析信息失败:[%s]\n", err.Error())
			//给前端回复错误信息
			response := messageString("error", "解析信息失败")
			conn.WriteMessage(messageType, response)
			continue
		}
		switch message.Type {
		case "getHostInfo":
			var hostStatus models.HostStatus
			//开启
			var wg sync.WaitGroup
			errChan := make(chan string, 4)
			getInfo := func(infoFunc func() interface{}, infoName string) {
				defer wg.Done()
				info := infoFunc()
				if info == nil {
					errChan <- fmt.Sprintf("获取%s信息失败", infoName)
					return
				}
				if infoName == "网络" {
					hostStatus.NetStatus = *(info.(*models.NetStatus))
				} else if infoName == "cpu" {
					hostStatus.HostBasicInfos.CpuInfo = *(info.(*models.HostItemStatu))
				} else if infoName == "内存" {
					hostStatus.HostBasicInfos.MemInfo = *(info.(*models.HostItemStatu))
				} else if infoName == "swap" {
					hostStatus.HostBasicInfos.SwapInfo = *(info.(*models.HostItemStatu))
				} else if infoName == "磁盘" {
					hostStatus.HostBasicInfos.DiskInfo = *(info.(*models.HostItemStatu))
				}
			}
			wg.Add(5)
			go getInfo(GetCpuInfo, "cpu")
			go getInfo(GetNetInfo, "网络")
			go getInfo(GetMemInfo, "内存")
			go getInfo(GetSwapInfo, "swap")
			go getInfo(GetDiskInfo, "磁盘")
			wg.Wait()
			close(errChan)
			if len(errChan) > 0 {
				for errMsg := range errChan {
					global.Log.Errorf("%s\n", errMsg)
					response := messageString("hostError", errMsg)
					conn.WriteMessage(messageType, response)
				}
			}
			response := messageString("hostInfo", hostStatus)
			conn.WriteMessage(messageType, response)
		case "cmdStdin":
			//执行命令
			HandleOrder_1(message.Data.(string), conn)
		case "ptyInfo":
			//设置pty的大小
			fmt.Printf("ptyInfo:%v %T\n", message.Data, message.Data)
			cols := message.Data.(map[string]interface{})["cols"].(float64)
			rows := message.Data.(map[string]interface{})["rows"].(float64)
			global.Log.Infof("Cols:%v %T Rows:%v %T\n", cols, cols, rows, rows)
			var ptyInfo models.PtyInfo = models.PtyInfo{
				Cols: uint16(cols),
				Rows: uint16(rows),
			}
			global.Bash.Cols = ptyInfo.Cols
			global.Bash.Rows = ptyInfo.Rows
			if global.Bash.CMD.Process == nil {
				fmt.Println("bash进程未启动")
				ptmx, err := pty.Start(global.Bash.CMD)
				global.Bash.Ptmx = ptmx
				if err != nil {
					global.Log.Errorf("启动bash进程以及伪终端pty失败: %s", err.Error())
					return
				}

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
			pty.Setsize(global.Bash.Ptmx, &pty.Winsize{
				Cols: ptyInfo.Cols,
				Rows: ptyInfo.Rows,
			})
			global.Log.Infof("Cols:%d Rows:%d,全局pty大小设置成功\n", ptyInfo.Cols, ptyInfo.Rows)
		}
	}
}
