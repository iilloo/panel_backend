package tests

import (
	"fmt"
	"panel_backend/global"
	"panel_backend/repository"
)

func DeleteUser(name string) {
	//删除用户
	err := repository.DeleteUserByName(name)
	if err != nil {
		global.Log.Error("删除用户失败")
		panic(err)
	}
	fmt.Println("删除用户成功")
	//用户数量清零
	global.UserCount = 0
}