package tests

import (
	"net/http"
	"panel_backend/global"
	"panel_backend/utils/jwts"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func RouteHello() gin.HandlerFunc {

	return func(c *gin.Context) {
		global.Log.Errorln("this is a error log")
		claims := jwts.UserClaims{
			Username: "admin",
			UserID:   1,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Second)),
			},
		}
		global.Secret = "6ec0358d-436c-4999-b001-837f03f85520"
		token := jwts.GenerateToken(claims)
		c.Header("Token", token)
		c.String(http.StatusOK, "Hello World!")
	}
}