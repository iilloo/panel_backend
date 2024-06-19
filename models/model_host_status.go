package models

type HostItemStatu struct {
	Name       string  `json:"name"`
	Percentage int8    `json:"percentage"`
	Current    float64 `json:"current"`
	Sum        float64 `json:"sum"`
	Suffix     string  `json:"suffix"`
}

type NetStatus struct {
	DownloadSpeed string `json:"downloadSpeed"`
	UploadSpeed   string `json:"uploadSpeed"`
	DownloadTotal string `json:"downloadTotal"`
	UploadTotal   string `json:"uploadTotal"`
}

type SysBasicInfo struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type HostBasicInfos struct {
	CpuInfo        HostItemStatu   `json:"cpuInfo"`
	MemInfo        HostItemStatu   `json:"memInfo"`
	SwapInfo       HostItemStatu   `json:"swapInfo"`
	DiskInfo       HostItemStatu   `json:"diskInfo"`
}

type HostStatus struct {
	HostBasicInfos HostBasicInfos `json:"hostBasicInfos"`
	NetStatus      NetStatus       `json:"netStatus"`
}
