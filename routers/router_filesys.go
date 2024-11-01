package routers

import (
	v1 "panel_backend/api/v1"

	"github.com/gin-gonic/gin"
)

func FileSysRouter(router *gin.Engine) {
	//文件系统相关路由

	// router.POST("/createUser", v1.CreateUser)
	// router.POST("/updateUser", v1.UpdateUser)
	// router.POST("/deleteUser", v1.DeleteUser)
	// router.POST("/getUser", v1.GetUser)
	r := router.Group("/fileSys")
	r.GET("/search", v1.SearchFile())
	r.POST("/add", v1.AddFile())
	r.POST("/delete", v1.DeleteFile())
	// r.DELETE("/delete", v1.DeleteFile())
	r.POST("/rename/:oldername/:newname", v1.RenameFile())
	r.GET("/read", v1.ReadFile())
	// r.POST("/read", v1.ReadFile())
	r.PUT("/write", v1.WriteFile())
	r.POST("/cutPaste", v1.CutPasteFile())
	r.GET("/copyPaste", v1.CopyPasteFile())
	r.POST("/uploadFile", v1.UploadFile())
	r.GET("/uploadFileProgress", v1.UploadFileProgress())

	r.POST("/uploadFolder", v1.UploadFolder())
	// r.GET("/uploadFolderProgress", v1.UploadFolderProgress())
	r.GET("/downloadFile", v1.DownloadFile())
	r.GET("/downloadFileGetSignature", v1.DownloadFileGetSignature())
}
