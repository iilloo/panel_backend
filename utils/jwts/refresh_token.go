package jwts

import (
	"panel_backend/global"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func RefreshToken(token *jwt.Token) (string, error) {
	// 刷新token
	claims := token.Claims.(*UserClaims)
	// 延长过期时间为30秒
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(2 * time.Hour))
	tokenString, err := token.SignedString([]byte(global.Secret))
	return tokenString, err
}
