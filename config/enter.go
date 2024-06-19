package config
// 
type Config struct {
	System System `yaml:"system"`
	Mysql  Mysql  `yaml:"mysql"`
	JWT    JWT    `yaml:"jwt"`
}

