package core

import (
	"fmt"
	_ "panel_backend/global"
	"panel_backend/repository"
)

// 前提是数据库初始化完成
func InitUserCount() int {
	//初始化用户数量
	users, err := repository.GetUserList()
	if len(users) == 0 || err != nil{
		fmt.Println("获取用户列表失败,用户数量是0")
		return 0
	} else {
		return 1
	}
}
