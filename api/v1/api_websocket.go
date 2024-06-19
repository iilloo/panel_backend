package v1

import (
	"panel_backend/services"

	"github.com/gin-gonic/gin"
)


func WS() gin.HandlerFunc{
	return func(c *gin.Context) {
		//websocket
		services.WS(c)
	}
}