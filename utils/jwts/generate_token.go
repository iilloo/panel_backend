package jwts

import (
	"os"
	"panel_backend/global"

	"github.com/golang-jwt/jwt/v5"
)


func GenerateToken(claims UserClaims) string {
	// 生成token（*jwt.Token类型）
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(global.Secret))
	if err != nil {
		global.Log.Error("jwt_token生成错误", err)
		os.Exit(1)
	}
	global.Log.Debugf("jwt_token生成成功[%s]\n", tokenString)
	return tokenString
}
	