package core

import (
	"os"
	"panel_backend/global"
)

func InitJwt() string{
	//初始化jwt
	secret := global.Config.JWT.Secret
	if secret == "" {
		global.Log.Error("jwt secret为空")
		os.Exit(1)
	}
	return secret
}