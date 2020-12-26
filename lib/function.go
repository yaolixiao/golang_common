package lib

import (
	"log"
	"flag"
	"os"
)

// 初始化函数：支持两种方式读取配置文件
// 1. 函数传入配置文件 Init("./conf/dev/")
// 2. 如果配置文件为空，则从命令行读取 -config conf/dev/
func Init(configPath string) error {
	return InitModule(configPath, []string{"base", "redis", "mysql"})
}

// 模块初始化
func InitModule(configPath string, modules []string) error {
	conf := flag.String("config", configPath, "input config file like ./conf/dev/")
	log.Printf("before Parse conf=%v *conf=%v\n", conf, *conf)
	flag.Parse()
	if *conf == "" {
		flag.Usage()
		os.Exit(1)
	}

	log.Println("========================================")
	log.Printf("[INFO] config=%s\n", *conf)
	log.Printf("[INFO] %s\n", "loaded config.")
}