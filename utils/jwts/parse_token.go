package jwts

import (
	"panel_backend/global"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func ParseToken(tokenString string) (*UserClaims, string , error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(global.Secret), nil
	})
	if err != nil {
		global.Log.Error("jwt_token解析错误", err)
		return nil, "", err
	}
	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		global.Log.Debugf("jwt_token解析成功\n")
		// global.Log.Infof("jwt_token解析成功[%s]\n", tokenString)
		// 刷新token
		now := time.Now()
		if claims.ExpiresAt.Time.Sub(now) < 10 * time.Minute {
			newTokenString, err := RefreshToken(token)
			if err != nil {
				global.Log.Error("jwt_token刷新错误", err)
				return nil, "", err
			}
			global.Log.Debugf("jwt_token刷新成功[%s]\n", tokenString)
			// global.Log.Infof("jwt_token刷新成功[%s]\n", tokenString)
			return claims, newTokenString, nil
		}
		// global.Log.Infof("jwt_token仍有效[%s]\n", tokenString)
		return claims, tokenString, nil
	}
	global.Log.Error("jwt_token无效")
	return nil, "",err
}