package models



// File 表示文件或目录
type File struct {
    Path    string    `json:"path"`     // 文件或目录的完整路径
    Name    string    `json:"name"`     // 文件或目录名称
    IsDir   bool      `json:"isDir"`    // 是否为目录
    Size    int64     `json:"size"`     // 文件大小(字节)
    ModTime string `json:"modTime"`  // 最后修改时间
}

type DeleteRequest struct {
	Path  string   `json:"path"`
	Names []string `json:"names"`
}

