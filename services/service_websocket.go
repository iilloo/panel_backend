package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"panel_backend/global"
	"panel_backend/utils/jwts"
	"sync"
	"time"

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
				return
			} else if websocket.IsCloseError(err, websocket.CloseGoingAway) {
				global.Log.Errorf("websocket连接被前端正常关闭\n")
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
				} else if infoName == "cpu"{
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
			global.Log.Infof("收到命令: %s\n", message.Data)
			HandleOrder(message.Data.(string), conn)
		}
		

	}
}
