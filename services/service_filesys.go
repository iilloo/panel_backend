package services

import (
	"archive/zip"
	"bufio"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	_ "io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"panel_backend/global"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"panel_backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func SearchFile(c *gin.Context) {
	//搜索
	path := c.Query("path")
	dirEntries, err := os.ReadDir(path)
	if err != nil {
		global.Log.Errorf("[%s]路径不合法:[%s]\n", path, err.Error())
		c.JSON(400, gin.H{
			"code": 400,
			"msg":  "路径不存在，请重新输入",
		})
		return
	}
	files := make([]*models.File, 0)
	for _, entry := range dirEntries {
		fileinfo, err := entry.Info()
		if err != nil {
			global.Log.Errorf("[%s]读取目录信息失败:[%s]\n", entry.Name(), err.Error())
			c.JSON(400, gin.H{
				"code": 400,
				"msg":  "读取文件信息失败",
			})
			return
		}
		path = strings.TrimRight(path, "/")
		file := models.File{
			Name:    fileinfo.Name(),
			Path:    path + "/" + fileinfo.Name(),
			Size:    fileinfo.Size(),
			ModTime: fileinfo.ModTime().Format("2006-01-02 15:04:05"),
			IsDir:   fileinfo.IsDir(),
		}
		files = append(files, &file)
	}
	// fmt.Printf("files: %v\n", files)
	global.Log.Debugf("返回[%s]信息成功\n", path)
	c.JSON(200, gin.H{
		"code":  200,
		"files": files,
	})
}
func AddFile(c *gin.Context) {
	var file models.File
	c.BindJSON(&file)
	path := file.Path
	name := file.Name
	isDir := file.IsDir
	// 去掉可能的最后一个斜杠
	path = strings.TrimRight(path, "/")
	if isDir {
		// 创建目录
		err := os.Mkdir(path+"/"+name, os.ModePerm)
		if err != nil {
			global.Log.Errorf("[%s]创建目录失败:[%s]\n", path+"/"+name, err.Error())
			c.JSON(500, gin.H{
				"code": 500,
				"msg":  "系统错误，创建目录失败",
			})
			return
		}
		global.Log.Debugf("创建[%s]目录成功\n", path+"/"+name)
		c.JSON(200, gin.H{
			"code": 200,
			"msg":  "创建成功",
		})
		return
	} else {
		// 创建文件
		_, err := os.Create(path + "/" + name)
		if err != nil {
			global.Log.Errorf("[%s]创建文件失败:[%s]\n", path+"/"+name, err.Error())
			c.JSON(500, gin.H{
				"msg": "系统错误，创建文件失败",
			})
			return
		}
		global.Log.Debugf("创建[%s]文件成功\n", path+"/"+name)
		c.JSON(200, gin.H{
			"code": 200,
			"msg":  "创建成功",
		})
		return
	}

}
func DeleteFile(c *gin.Context) {
	// body, _ := io.ReadAll(c.Request.Body)
	// fmt.Printf("消息体：%s",string(body))
	if c.GetHeader("Override") != "DELETE" {
		global.Log.Errorf("删除文件请求不合法,Override请求头缺失或有误\n")
		c.JSON(400, gin.H{
			"code": 400,
			"msg":  "请求不合法",
		})
		return
	}
	var deleteRequest models.DeleteRequest
	c.BindJSON(&deleteRequest)
	path := strings.TrimRight(deleteRequest.Path, "/")
	// 循环删除
	for _, name := range deleteRequest.Names {
		currentPath := path + "/" + name
		currentName := strings.Split(name, ".")
		if len(currentName) == 1 {
			err := os.RemoveAll(currentPath)
			if err != nil {
				global.Log.Errorf("[%s]删除失败:[%s]\n", currentPath, err.Error())
				c.JSON(500, gin.H{
					"code": 500,
					"msg":  "系统错误，删除失败",
				})
				return
			}
			global.Log.Debugf("删除[%s]成功\n", currentPath)
		} else {
			err := os.Remove(currentPath)
			if err != nil {
				global.Log.Errorf("[%s]删除失败:[%s]\n", currentPath, err.Error())
				c.JSON(500, gin.H{
					"code": 500,
					"msg":  "系统错误，删除失败",
				})
				return
			}
			global.Log.Debugf("删除[%s]成功\n", currentPath)
		}
	}
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "删除成功",
	})
}
func RenameFile(c *gin.Context) {
	type RenameRequest struct {
		Path string `json:"path"`
	}
	var renameRequest RenameRequest
	c.BindJSON(&renameRequest)
	path := renameRequest.Path
	name := c.Param("oldername")
	newName := c.Param("newname")
	path = strings.TrimRight(path, "/")
	err := os.Rename(path+"/"+name, path+"/"+newName)
	if err != nil {
		global.Log.Errorf("[%s]重命名失败:[%s]\n", path+"/"+name, err.Error())
		c.JSON(500, gin.H{
			"code": 500,
			"msg":  "系统错误，重命名失败",
		})
		return
	}
	global.Log.Debugf("重命名[%s]成功\n", path+"/"+name)
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "重命名成功",
	})
}
func ReadFile(c *gin.Context) {
	// 读取文件内容
	path := c.Query("path")
	global.Log.Debugf("读取[%s]文件\n", path)
	done := make(chan bool, 1)
	cancel := make(chan bool, 1)
	go func() {
		file, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
		if err != nil {
			global.Log.Errorf("[%s]打开文件失败:[%s]\n", path, err.Error())
			c.JSON(500, gin.H{
				"msg": "系统错误，打开文件失败",
			})
			return
		}
		defer file.Close()

		reader := bufio.NewReader(file)
		content := ""
		buf := make([]byte, 1024)
	lable1:
		for {
			select {
			case <-cancel:
				global.Log.Errorf("读取[%s]文件取消\n", path)
				return
			default:
				n, err := reader.Read(buf)
				if err != nil {
					break lable1
				}
				content += string(buf[:n])
			}
		}
		global.Log.Debugf("读取[%s]文件成功\n", path)
		c.JSON(200, gin.H{
			"text": content,
		})
		done <- true
	}()
	select {
	case <-done:
		return
	case <-time.After(3 * time.Second):
		global.Log.Errorf("读取[%s]文件超时\n", path)
		// 通知协程取消
		cancel <- true
		c.JSON(500, gin.H{
			"msg": "系统错误或文件过大，读取文件超时",
		})
		return
	}
}

type WriteFileText struct {
	Path string `json:"path"`
	Text string `json:"text"`
	Name string `json:"name"`
}

func WriteFile(c *gin.Context) {
	// 写入文件内容
	var writeFileText WriteFileText
	c.BindJSON(&writeFileText)
	path := writeFileText.Path
	path = strings.TrimRight(path, "/") + "/" + writeFileText.Name

	text := writeFileText.Text
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		global.Log.Errorf("[%s]打开文件失败:[%s]\n", path, err.Error())
		c.JSON(500, gin.H{
			"msg": "系统错误，打开文件失败",
		})
		return
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	_, err = writer.Write([]byte(text))
	if err != nil {
		global.Log.Errorf("[%s]写入文件失败:[%s]\n", path, err.Error())
		c.JSON(500, gin.H{
			"msg": "系统错误，写入文件失败",
		})
		return
	}
	// 将缓冲区中的数据刷新到磁盘
	err = writer.Flush()
	if err != nil {
		global.Log.Errorf("[%s]刷新缓冲区失败:[%s]\n", path, err.Error())
		c.JSON(500, gin.H{
			"msg": "系统错误，刷新缓冲区失败",
		})
		return
	}
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "写入成功",
	})

}

type PasteRequest struct {
	OldPath     string   `json:"oldPath"`
	NewPath     string   `json:"newPath"`
	Names       []string `json:"names"`
	DeleteNames []string `json:"deleteNames"`
}

func CutPasteFile(c *gin.Context) {
	// 粘贴文件
	var pasteRequest PasteRequest
	c.BindJSON(&pasteRequest)
	oldPath := pasteRequest.OldPath
	newPath := pasteRequest.NewPath
	names := pasteRequest.Names
	// deleteNames := pasteRequest.DeleteNames
	global.Log.Infof("oldPath: %s, newPath: %s\n, names: %v\n", oldPath, newPath, names)
	oldPath = strings.TrimRight(oldPath, "/")
	newPath = strings.TrimRight(newPath, "/")
	// 删除文件
	// for _, name := range deleteNames {
	// 	currentPath := filepath.Join(newPath, name)
	// 	fileInfo, err := os.Stat(currentPath)
	// 	if err != nil {
	// 		global.Log.Errorf("[%s]获取文件信息失败:[%s]\n", currentPath, err.Error())
	// 		c.JSON(500, gin.H{
	// 			"code": 500,
	// 			"msg":  "系统错误，获取欲覆盖文件信息失败",
	// 		})
	// 		return
	// 	}
	// 	if fileInfo.IsDir() {
	// 		err := os.RemoveAll(currentPath)
	// 		if err != nil {
	// 			global.Log.Errorf("[%s]删除失败:[%s]\n", currentPath, err.Error())
	// 			c.JSON(500, gin.H{
	// 				"code": 500,
	// 				"msg":  "系统错误，删除失败",
	// 			})
	// 			return
	// 		}
	// 	} else {
	// 		err := os.Remove(currentPath)
	// 		if err != nil {
	// 			global.Log.Errorf("[%s]删除失败:[%s]\n", currentPath, err.Error())
	// 			c.JSON(500, gin.H{
	// 				"code": 500,
	// 				"msg":  "系统错误，删除失败",
	// 			})
	// 			return
	// 		}
	// 	}
	// }

	//移动文件
	for _, name := range names {
		err := os.Rename(oldPath+"/"+name, newPath+"/"+name)
		if err != nil {
			global.Log.Errorf("[%s]移动失败:[%s]\n", oldPath+"/"+name, err.Error())
			c.JSON(500, gin.H{
				"code": 500,
				"msg":  "系统错误，移动失败",
			})
			return
		}
		global.Log.Debugf("移动[%s]成功\n", oldPath+"/"+name)
	}
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "移动成功",
	})
}

// FolderSize 计算文件夹大小
//
//	func FolderSize(path string) int64 {
//		var size int64 = 0
//		entries, err := os.ReadDir(path)
//		if err != nil {
//			global.Log.Errorf("[%s]路径不合法:[%s]\n", path, err.Error())
//			return -1
//		}
//		for _, entry := range entries {
//			fileInfo, err := entry.Info()
//			if err != nil {
//				global.Log.Errorf("[%s]获取文件信息失败:[%s]\n", entry.Name(), err.Error())
//				return -1
//			}
//			if fileInfo.IsDir() {
//				size += FolderSize(path + "/" + entry.Name())
//			} else {
//				size += fileInfo.Size()
//			}
//		}
//		return size
//	}
func FolderSize(root string) (uint64, error) {
	var size uint64

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// 如果遇到权限问题或其他错误，记录错误但继续遍历
			global.Log.Errorf("访问路径出错 %s: %v\n", path, err)
			return filepath.SkipDir
		}

		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				global.Log.Errorf("无法获取文件信息 %s: %v\n", path, err)
				return nil
			}
			atomic.AddUint64(&size, uint64(info.Size()))
		}

		return nil
	})

	if err != nil {
		return size, fmt.Errorf("遍历目录时发生错误: %w", err)
	}

	return size, nil
}

// AllFilesSize 计算所有文件大小
func AllFilesSize(path string, names []string) uint64 {
	var size uint64
	for _, name := range names {
		fileFullPath := path + "/" + name
		fileInfo, err := os.Stat(fileFullPath)
		if err != nil {
			global.Log.Errorf("[%s]获取文件信息失败:[%s]\n", fileFullPath, err.Error())
			continue
		}
		if fileInfo.IsDir() {
			ssize, _ := FolderSize(fileFullPath)
			size += ssize
		} else {
			size += uint64(fileInfo.Size())
		}
	}
	return size
}

// copyFileWithProgress 拷贝文件并发送进度
func copyFileWithProgress(source, destination string, progressChan chan<- uint64, c *gin.Context) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		global.Log.Errorf("[%s]打开文件失败:[%s]\n", source, err.Error())
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(destination)
	if err != nil {
		global.Log.Errorf("[%s]创建文件失败:[%s]\n", destination, err.Error())
		return err
	}
	defer destinationFile.Close()
	fmt.Fprintf(c.Writer, "data: Copying %s to %s\n\n", source, destination)
	fmt.Fprintf(c.Writer, "data: SrcFileName: %s\n\n", source)
	fmt.Fprintf(c.Writer, "data: DestFileName: %s\n\n", destination)
	c.Writer.(http.Flusher).Flush()
	global.Log.Infof("当前正在进行Copying %s to %s\n", source, destination)
	copiedBytes := uint64(0)
	buffer := make([]byte, 1024*1024) // 1MB buffer

	for {
		n, err := sourceFile.Read(buffer)
		if err != nil && err != io.EOF {
			global.Log.Errorf("[%s]读取文件失败:[%s]\n", source, err.Error())
			return err
		}
		if n == 0 {
			break
		}

		_, err = destinationFile.Write(buffer[:n])
		if err != nil {
			global.Log.Errorf("[%s]写入文件失败:[%s]\n", destination, err.Error())
			return err
		}

		copiedBytes = uint64(n)
		progressChan <- copiedBytes
	}

	return nil
}

// copyDirWithProgress 递归拷贝目录并发送进度
func copyDirWithProgress(source, destination string, progressChan chan<- uint64, c *gin.Context) error {
	entries, err := os.ReadDir(source)
	if err != nil {
		global.Log.Errorf("[%s]读取目录失败:[%s]\n", source, err.Error())
		return err
	}

	if err := os.MkdirAll(destination, os.ModePerm); err != nil {
		global.Log.Errorf("[%s]创建目录失败:[%s]\n", destination, err.Error())
		return err
	}

	for _, entry := range entries {
		sourcePath := filepath.Join(source, entry.Name())
		destinationPath := filepath.Join(destination, entry.Name())

		var err error
		if entry.IsDir() {
			err = copyDirWithProgress(sourcePath, destinationPath, progressChan, c)
		} else {
			err = copyFileWithProgress(sourcePath, destinationPath, progressChan, c)
		}

		if err != nil {
			global.Log.Errorf("[%s]拷贝失败:[%s]\n", sourcePath, err.Error())
			return err
		}
	}
	return nil
}

func CopyPasteFile(c *gin.Context) {
	//设置SSE http长连接响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	// 粘贴文件
	oldPath := c.Query("oldPath")
	newPath := c.Query("newPath")
	names := c.Query("names")
	var fileNames []string
	if err := json.Unmarshal([]byte(names), &fileNames); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "存在不合法的文件名",
		})
		return
	}

	global.Log.Infof("oldPath: %s, newPath: %s\n, names: %v\n", oldPath, newPath, fileNames)
	oldPath = strings.TrimRight(oldPath, "/")
	newPath = strings.TrimRight(newPath, "/")
	// 计算所有文件大小
	totalBytes := AllFilesSize(oldPath, fileNames)
	global.Log.Infof("totalBytes: %d\n", totalBytes)
	// 通知前端文件大小
	fmt.Fprintf(c.Writer, "data: TotalBytes: %d\n\n", totalBytes)
	c.Writer.(http.Flusher).Flush()

	progressChan := make(chan uint64)
	doneChan := make(chan bool)
	// 开启协程监测进度
	go func() {
		var copiedBytes uint64 = 0
		var preProgressPercentage int = -1
		for progress := range progressChan {
			copiedBytes += progress
			progressPercentage := int(float64(copiedBytes) / float64(totalBytes) * 100)
			if progressPercentage != preProgressPercentage {
				preProgressPercentage = progressPercentage
				global.Log.Infof("copied: %d当前progressPercentage: %d\n", copiedBytes, progressPercentage)
				fmt.Fprintf(c.Writer, "data: Percent: %d\n\n", progressPercentage)
				// 刷新缓冲区，确保数据立即发送
				if flusher, ok := c.Writer.(http.Flusher); ok {
					if flusher != nil {
						flusher.Flush()
					} else {
						global.Log.Errorf("flusher is nil")
					}
				} else {
					global.Log.Errorf("c.Writer does not implement http.Flusher")
				}
			} else {
				continue
			}
			if copiedBytes == totalBytes {
				doneChan <- true
				break
			}
		}
	}()
	// 复制文件
	for _, name := range fileNames {
		destPath := filepath.Join(newPath, name)
		srcPath := filepath.Join(oldPath, name)
		var err error
		fileInfo, err := os.Stat(srcPath)
		if err != nil {
			global.Log.Errorf("[%s]获取文件信息失败:[%s]\n", srcPath, err.Error())
			return
		}
		if fileInfo.IsDir() {
			err = copyDirWithProgress(srcPath, destPath, progressChan, c)
		} else {
			err = copyFileWithProgress(srcPath, destPath, progressChan, c)
		}

		if err != nil {
			global.Log.Errorf("[%s]拷贝失败:[%s]\n", srcPath, err.Error())
			return
		}
	}
	close(progressChan)
	// 等待拷贝完成
	if <-doneChan {
		fmt.Fprintf(c.Writer, "data: Copy operation completed!\n\n")
		c.Writer.(http.Flusher).Flush()
	}
}

//	type UploadProgress struct {
//		copied map[string]int
//		mu    sync.Mutex
//	}
//
//	func init() {
//		uprogress := &UploadProgress{
//			copied: make(map[string]int),
//		}
//	}
var uploadCopied sync.Map
var uploadTotal sync.Map
var uploadFileCount sync.Map
var uploadFileName sync.Map

var uploadDone sync.Map

// 上传单个文件，并记录当前上传文件名，和已上传大小
func UploadFileWithProgress(resFile *multipart.FileHeader, destPath string, index string) error {
	// 上传文件
	src, err := resFile.Open()
	if err != nil {
		global.Log.Errorf("[%s]打开文件失败:[%s]\n", resFile.Filename, err.Error())
		return err
	}
	if ch, ok := uploadFileName.Load(index); ok {
		global.Log.Infof("记录当前上传文件名%s成功\n", resFile.Filename)
		// 将interface{}断言为chan string类型
		chanStr := ch.(chan string)
		chanStr <- resFile.Filename
	}
	defer src.Close()
	dst, err := os.Create(destPath)
	if err != nil {
		global.Log.Errorf("[%s]创建文件失败:[%s]\n", destPath, err.Error())
		return err
	}
	defer dst.Close()
	buf := make([]byte, 1024*1024) // 1MB buffer
	for {
		n, err := src.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			global.Log.Errorf("[%s]读取文件失败:[%s]\n", resFile.Filename, err.Error())
			return err
		}
		_, err = dst.Write(buf[:n])
		if err != nil {
			global.Log.Errorf("[%s]写入文件失败:[%s]\n", destPath, err.Error())
			return err
		}
		// 记录上传进度
		if copied, ok := uploadCopied.Load(index); ok {
			uploadCopied.Store(index, uint64(n)+copied.(uint64))
		} else {
			uploadCopied.Store(index, uint64(n))
		}
	}

	return nil
}

// UploadFile 上传文件
func UploadFile(c *gin.Context) {
	form, _ := c.MultipartForm()
	files := form.File["files"]
	path := c.PostForm("path")
	index := c.PostForm("timeIndex")
	//声明一个管道，用于存储当前上传文件的名称
	var fileNamesChan = make(chan string, len(files))
	uploadFileName.Store(index, fileNamesChan)

	//存储文件的个数
	uploadFileCount.Store(index, len(files))

	//声明一个管道，用于存储当前上传的完成状态
	var doneChan = make(chan bool, 1)
	uploadDone.Store(index, doneChan)

	// 去掉可能的最后一个斜杠
	// path = strings.TrimRight(path, "/")
	totalSize := uint64(0)
	for _, file := range files {
		totalSize += uint64(file.Size)
	}
	uploadTotal.Store(index, totalSize)
	for _, file := range files {
		// 使用filepath.Clean()去掉路径中的多余斜杠
		dst := filepath.Join(filepath.Clean(path), file.Filename)
		//err := c.SaveUploadedFile(file, dst)
		err := UploadFileWithProgress(file, dst, index)
		if err != nil {
			c.JSON(500, gin.H{
				"code": 500,
				"msg":  fmt.Sprintf("系统错误，%s上传失败", file.Filename),
			})
		}
	}
	// 上传完成向相应的donechan中存入true
	if ch, ok := uploadDone.Load(index); ok {
		doneChan := ch.(chan bool)
		doneChan <- true
	}

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "上传成功",
	})
}

// 上传文件夹
func UploadFolder(c *gin.Context) {
	// 获取path、index等信息
	path := c.PostForm("path")
	index := c.PostForm("timeIndex")
	// 获取上传的文件
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取文件夹名称
	folders := form.Value["folders"]
	filesCount := 0
	totalSize := uint64(0)
	for _, folderName := range folders {
		key := fmt.Sprintf("files[%s]", folderName)
		files := form.File[key]
		filesCount += len(files)
		for _, file := range files {
			totalSize += uint64(file.Size)
		}
	}
	uploadTotal.Store(index, totalSize)
	//声明一个管道，用于存储当前上传文件的名称
	var fileNamesChan = make(chan string, filesCount)
	uploadFileName.Store(index, fileNamesChan)

	//存储文件的个数
	uploadFileCount.Store(index, filesCount)

	//声明一个管道，用于存储当前上传的完成状态
	var doneChan = make(chan bool, 1)
	uploadDone.Store(index, doneChan)

	// 遍历所有文件夹名称
	for _, folderName := range folders {
		// 创建文件夹路径
		folderPath := filepath.Join(filepath.Clean(path), folderName)
		if _, err := os.Stat(folderPath); os.IsNotExist(err) {
			os.MkdirAll(folderPath, os.ModePerm) // 创建文件夹
		}

		// 获取当前文件夹下的所有文件
		key := fmt.Sprintf("files[%s]", folderName)
		files := form.File[key]

		// 保存文件到服务器
		for _, file := range files {
			dst := filepath.Join(folderPath, file.Filename)
			err := UploadFileWithProgress(file, dst, index)
			if err != nil {
				c.JSON(500, gin.H{
					"code": 500,
					"msg":  fmt.Sprintf("系统错误，%s上传失败", file.Filename),
				})
			}
		}
	}
	// 上传完成向相应的donechan中存入true
	if ch, ok := uploadDone.Load(index); ok {
		doneChan := ch.(chan bool)
		doneChan <- true
	}

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "上传成功",
	})

}

func UploadFileProgress(c *gin.Context) {
	//设置SSE http长连接响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	index := c.Query("timeIndex")
	// 返回总大小
	var tSize uint64
	for {
		if totalSize, ok := uploadTotal.Load(index); ok {
			global.Log.Infof("index:%s-totalSize: %d\n", index, totalSize.(uint64))
			fmt.Fprintf(c.Writer, "data: TotalSize: %d\n\n", totalSize.(uint64))
			// 刷新缓冲区，确保数据立即发送
			c.Writer.(http.Flusher).Flush()
			uploadTotal.Delete(index)
			tSize = totalSize.(uint64)
			global.Log.Infof("tSize: %d\n", tSize)
			break
		} else {
			time.Sleep(10 * time.Millisecond)
		}
	}
	global.Log.Infof("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\n")
	// 开启一个协程返回当前上传文件名
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		var count int = 0
		if s, ok := uploadFileCount.Load(index); ok {
			count = s.(int)
		} else {
			global.Log.Errorf("uploadFileCount.Load(index)失败\n")
		}
		if ch, ok := uploadFileName.Load(index); ok {
			// 将interface{}断言为chan string类型
			chanStr := ch.(chan string)
			var i int = 0
			for name := range chanStr {
				i++
				fmt.Fprintf(c.Writer, "data: FileName: %s\n\n", name)
				// 刷新缓冲区，确保数据立即发送
				c.Writer.(http.Flusher).Flush()
				global.Log.Infof("name: %s\n", name)
				if i == count {
					break
				}
			}
		} else {
			global.Log.Errorf("uploadFileName.Load(index)失败\n")
		}
	}()

	// 返回上传进度
	var preCopied uint64 = 0
	var preProgressPercentage int = 0
	for {
		if copied, ok := uploadCopied.Load(index); ok {
			tmp := copied.(uint64)
			if tmp == preCopied {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			preCopied = tmp
			progressPercentage := int(float64(tmp) / float64(tSize) * 100)
			if progressPercentage != preProgressPercentage {
				preProgressPercentage = progressPercentage
				global.Log.Infof("percent: %d\n", progressPercentage)
				fmt.Fprintf(c.Writer, "data: progressPercentage: %d\n\n", progressPercentage)
				// 刷新缓冲区，确保数据立即发送
				c.Writer.(http.Flusher).Flush()

			}
			// 上传完成
			if copied.(uint64) == tSize {
				// 删除对应的uploadCopied map中对应的数据
				uploadCopied.Delete(index)
				break
			}
		}
	}

	if ch, ok := uploadDone.Load(index); ok {
		doneChan := ch.(chan bool)
		//阻塞在这里，直到上传完成
		if <-doneChan {
			// 这里可能存在的问题：返回给前端已complete但上面协程的返回文件名还未完成
			// fmt.Fprintf(c.Writer, "data: Upload operation completed!\n\n")
			global.Log.Infof("cccccccccccccccccccccccccccccccccc\n")
			// c.Writer.(http.Flusher).Flush()
			// if ch, ok := uploadFileName.Load(index); ok {
			// 	chanStr := ch.(chan string)
			// 	close(chanStr)
			// 	uploadFileName.Delete(index)
			// }
		}
		// 删除对应的doneChan，防止内存泄漏，删除map中对应的数据
		close(doneChan)
		uploadDone.Delete(index)
	}
	// 等待wg.Wait()
	wg.Wait()
	//删除对应的fileName管道，删除map中对应的数据
	if ch, ok := uploadFileName.Load(index); ok {
		chanStr := ch.(chan string)
		close(chanStr)
		uploadFileName.Delete(index)
	}
	//删除对应的uploadFileCount map中对应的数据
	uploadFileCount.Delete(index)
	// 等待上传文件名完全返回后再返回complete，之后前端关闭SSE连接
	fmt.Fprintf(c.Writer, "data: Upload operation completed!\n\n")
	c.Writer.(http.Flusher).Flush()
	global.Log.Infof("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb\n")
}

// 以下为下载文件的代码
type fileInfo struct {
	Name  string `json:"name"`
	IsDir bool   `json:"isDir"`
}

// type downloadFileInfo struct {
// 	Path  string   `json:"path"`
// 	FilesInfo []fileInfo `json:"filesInfo"`
// }

func singleFileDownload(c *gin.Context, path string, name string) {
	// 打开文件
	fileFullPath := filepath.Join(path, name)
	file, err := os.Open(fileFullPath)
	if err != nil {
		global.Log.Errorf("Failed to open file: %s, error: %v", fileFullPath, err)
		c.JSON(500, gin.H{
			"code": 500,
			"msg":  "File not found",
		})
		return
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		global.Log.Errorf("Failed to get file info: %v", err)
		c.String(http.StatusInternalServerError, "Error retrieving file info")
		return
	}
	global.Log.Infof("下载的文件的size：%d\n", fileInfo.Size())
	global.Log.Infof("下载的文件的名字：%v\n", name)
	// 设置响应头，告诉浏览器是文件下载
	c.Writer.Header().Set("Need-ResponseHeader", "true")
	// // 对文件名进行 URL 编码, 确保浏览器能正确解析中文文件名
	encodedName := url.QueryEscape(filepath.Base(fileFullPath))
	c.Writer.Header().Set("Content-Disposition", "attachment; filename="+encodedName)
	c.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	c.Writer.Header().Set("Content-Type", "application/octet-stream")
	// 流式传输文件
	if _, err := io.Copy(c.Writer, file); err != nil {
		global.Log.Printf("Failed to copy file to response: %v", err)
	}
}

func addFileToZip(zipWriter *zip.Writer, fileFullPath string, prefix string) error {
	// 打开文件
	file, err := os.Open(fileFullPath)
	if err != nil {
		global.Log.Errorf("Failed to open file: %s, error: %v", fileFullPath, err)
		return err
	}
	defer file.Close()
	global.Log.Infof("prefix: %s\n", prefix) ////////////////////////////
	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	// 创建一个文件头
	header, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		return err
	}

	// 设置文件头中的文件名
	global.Log.Infof("fileName: %s\n", fileInfo.Name()) ////////////////////////////
	header.Name = fileInfo.Name()

	// 写入文件头
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	// 写入文件内容
	_, err = io.Copy(writer, file)
	if err != nil {
		return err
	}

	return nil
}

func addDirToZip(zipWriter *zip.Writer, dirPath string, prefix string) error {
	// 读取目录下的所有文件
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		entryPath := filepath.Join(dirPath, entry.Name())
		if entry.IsDir() {
			// 如果是目录，递归添加目录下的所有文件
			err = addDirToZip(zipWriter, entryPath, prefix+entry.Name()+"/")
		} else {
			// 如果是文件，直接添加
			err = addFileToZip(zipWriter, entryPath, prefix)
		}
		if err != nil {
			return err
		}
	}

	return nil
}
func multipleFilesDownload(c *gin.Context, path string, filesInfo []fileInfo) {
	// 在path目录下创建一个临时文件夹download
	tempDir, err := os.MkdirTemp(path, "download")
	global.Log.Infof("tempDir: %s\n", tempDir) ////////////////////////////
	if err != nil {
		global.Log.Errorf("Failed to create temp dir: %v", err)
		c.JSON(500, gin.H{
			"code": 500,
			"msg":  "Failed to create temp dir",
		})
		return
	}
	// defer os.RemoveAll(tempDir)

	// 创建一个 zip 文件
	zipFileName := filepath.Join(tempDir, "download.zip")
	zipFile, err := os.Create(zipFileName)
	global.Log.Infof("zipFileName: %s\n", zipFileName) ////////////////////////////
	if err != nil {
		global.Log.Errorf("Failed to create zip file: %v", err)
		c.JSON(500, gin.H{
			"code": 500,
			"msg":  "Failed to create zip file",
		})
		return
	}
	// defer zipFile.Close()

	// 创建 zip writer
	zipWriter := zip.NewWriter(zipFile)
	// defer zipWriter.Close()

	// 将文件添加到 zip 文件
	for _, fileInfo := range filesInfo {
		fileFullPath := filepath.Join(path, fileInfo.Name)
		if fileInfo.IsDir {
			// 如果是目录，递归添加目录下的所有文件
			err = addDirToZip(zipWriter, fileFullPath, "")
		} else {
			// 如果是文件，直接添加
			global.Log.Infof("是文件类型\n") /////////////////////////////
			err = addFileToZip(zipWriter, fileFullPath, "")
		}
		if err != nil {
			global.Log.Errorf("Failed to add file to zip: %v", err)
			c.JSON(500, gin.H{
				"code": 500,
				"msg":  "Failed to add file to zip",
			})
			return
		}
	}
	zipWriter.Close()
	zipFile.Close()
	zip_file, _ := os.Open(zipFileName)
	fileInfo, _ := zip_file.Stat()
	zipSize := fileInfo.Size()
	global.Log.Infof("zipFileSize: %v\n", zipSize) /////////////////////////////
	c.Writer.Header().Set("Content-Disposition", "attachment; filename=download.zip")
	c.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", zipSize))
	c.Writer.Header().Set("Content-Type", "application/octet-stream")

	// 流式传输文件
	if _, err := io.Copy(c.Writer, zip_file); err != nil {
		global.Log.Errorf("Failed to copy file to response: %v", err)
	}
	zip_file.Close()
	os.RemoveAll(tempDir)
	
}
func GenerateHMACSHA256(data, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func DownloadFileGetSignature(c *gin.Context) {

	// expiration := time.Now().Add(10 * time.Minute).Unix()
	// newUUID := uuid.New().String()
	// // 生成签名
	// signature := GenerateHMACSHA256(dataToSign, newUUID)

}

var expiration int64 = 0
var newUUID string = ""
var token string = ""

func DownloadFile(c *gin.Context) {
	global.Log.Infof("进入DownloadFile服务\n")
	isGetSignature := c.Query("getSinature")
	if isGetSignature == "true" {
		newUUID = uuid.New().String()
		token = c.Query("token")
		expiration = time.Now().Add(10 * time.Minute).Unix()
		global.Log.Infof("expiration: %d\n", expiration)

		// 生成签名
		signature := GenerateHMACSHA256(token, newUUID)
		c.JSON(200, gin.H{
			"code":      200,
			"signature": signature,
		})
		return
	}
	// global.Log.Infof("-------------------------DownloadFile\n")
	// global.Log.Infof("c.Request.Body: %v\n", c.Request.Body)
	// 获取 path 参数
	path := c.Query("path")
	// 获取 filesInfo 参数（JSON 字符串）
	filesInfoStr := c.Query("filesInfo")
	// 获取 token 参数
	signature := c.Query("signature")
	// 验证 token
	expectedSignature := GenerateHMACSHA256(token, newUUID)
	nowTime := time.Now().Unix()
	if signature != expectedSignature || nowTime > expiration {
		global.Log.Errorf("signature:%s,expectedSignature:%s\n", signature, expectedSignature)
		global.Log.Errorf("expiration:%d,nowTime:%d\n", expiration, nowTime)
		c.JSON(401, gin.H{
			"code": 401,
			"msg":  "无权限下载文件",
		})
		return
	}

	// 定义存储解析后数据的切片
	var filesInfo []fileInfo

	// 解析 JSON 字符串为 FileInfo 结构体
	if err := json.Unmarshal([]byte(filesInfoStr), &filesInfo); err != nil {
		c.JSON(400, gin.H{
			"code": 400,
			"msg":  "欲下载文件信息后端解析失败",
		})
		return
	}
	global.Log.Infof("path: %s, filesInfo: %s\n", path, filesInfoStr)
	// var downloadFileInfo downloadFileInfo

	global.Log.Debugf("下载[%s]文件,文件为[%v]\n", path, filesInfo)
	if len(filesInfo) == 1 && !filesInfo[0].IsDir {
		// 单个文件，直接传输
		singleFileDownload(c, path, filesInfo[0].Name)
	} else if len(filesInfo) > 1 || (len(filesInfo) == 1 && filesInfo[0].IsDir) {
		// 多个文件，打包成 zip 压缩包
		global.Log.Infof("多个文件\n")
		multipleFilesDownload(c, path, filesInfo)

	} else {
		c.JSON(400, gin.H{
			"code": 400,
			"msg":  "下载文件为空",
		})
	}
}
