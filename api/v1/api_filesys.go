package v1

import (
	"panel_backend/services"

	"github.com/gin-gonic/gin"
)

func SearchFile() gin.HandlerFunc{
	return func(c *gin.Context) {
		//搜索某个目录下的文件
		services.SearchFile(c)
	}
}

func AddFile() gin.HandlerFunc{
	return func(c *gin.Context) {
		//添加文件或文件夹
		services.AddFile(c)
	}
}

func DeleteFile() gin.HandlerFunc{
	return func(c *gin.Context) {
		//删除文件或文件夹
		services.DeleteFile(c)
	}
}

func RenameFile() gin.HandlerFunc{
	return func(c *gin.Context) {
		//重命名文件或文件夹
		services.RenameFile(c)
	}
}

func ReadFile() gin.HandlerFunc{
	return func(c *gin.Context) {
		//读取文件内容
		services.ReadFile(c)
	}
}

func WriteFile() gin.HandlerFunc{
	return func(c *gin.Context) {
		//写入文件内容
		services.WriteFile(c)
	}
}

func CutPasteFile() gin.HandlerFunc{
	return func(c *gin.Context) {
		//粘贴文件
		services.CutPasteFile(c)
	}
}

func CopyPasteFile() gin.HandlerFunc{
	return func(c *gin.Context) {
		//复制文件
		services.CopyPasteFile(c)
	}
}

func UploadFile() gin.HandlerFunc{
	return func(c *gin.Context) {
		//上传文件
		services.UploadFile(c)
	}
}

func UploadFolder() gin.HandlerFunc{
	return func(c *gin.Context) {
		//上传文件夹
		services.UploadFolder(c)
	}
}

func UploadFileProgress() gin.HandlerFunc{
	return func(c *gin.Context) {
		//获取上传文件进度
		services.UploadFileProgress(c)
	}
}