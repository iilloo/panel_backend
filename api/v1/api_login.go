package v1

import (
	_"fmt"
	_"panel_backend/models"
	"panel_backend/services"
	_"panel_backend/utils/jwts"
	_"time"

	"github.com/gin-gonic/gin"
	_"github.com/golang-jwt/jwt/v5"
)

func Login() gin.HandlerFunc{
	return func(c *gin.Context) {
		services.Login(c)
	}
}