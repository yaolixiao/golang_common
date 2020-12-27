package lib

import (
	"log"
	"flag"
	"os"
	"net"
	"fmt"
	"time"
)

var TimeLocation *time.Location
var TimeFormat = "2006-01-02 15:04:05"
var LocalIP = net.ParseIP("127.0.0.1")

// 初始化函数：支持两种方式读取配置文件
// 1. 函数传入配置文件 Init("./conf/dev/")
// 2. 如果配置文件为空，则从命令行读取 -config conf/dev/
// 3. 命令行读取优先级大于函数传入
func Init(configPath string) error {
	return InitModule(configPath, []string{"base", "redis", "mysql"})
}

// 模块初始化
func InitModule(configPath string, modules []string) error {
	conf := flag.String("config", configPath, "input config file like ./conf/dev/")
	flag.Parse()
	if *conf == "" {
		flag.Usage()
		os.Exit(1)
	}

	log.Println("========================================")
	log.Printf("[INFO] config=%s\n", *conf)
	log.Printf("[INFO] %s\n", "start loading config.")

	// 优先设置IP，便于日志打印
	ips := getLocationIPs()
	if len(ips) > 0 {
		LocalIP = ips[0]
	}

	// 解析配置文件目录
	if err := ParseConfPath(*conf); err != nil {
		return err
	}

	// 初始化配置文件
	if err := InitViperConf(); err != nil {
		return err
	}

	// 加载base配置
	if InArrayString("base", modules) {
		if err := InitBaseConf(GetConfPath("base")); err != nil {
			fmt.Printf("[ERROR] %s %s\n", time.Now().Format(TimeFormat), "InitBaseConf:" + err.Error())
		}
	}

	// 加载redis配置
	if InArrayString("redis", modules) {
		if err := InitRedisConf(GetConfPath("redis_map")); err != nil {
			fmt.Printf("[ERROR] %s %s\n", time.Now().Format(TimeFormat), "InitRedisConf:" + err.Error())
		}
	}

	// 加载mysql配置 todo

	// 设置时区
	if location, err := time.LoadLocation(ConfBase.TimeLocation); err != nil {
		return err
	} else {
		TimeLocation = location
	}

	log.Println("[INFO] success loading config.")
	return nil
}

// 获取本地IP
func getLocationIPs() (ips []net.IP) {
	interfaceAddrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}

	for _, address := range interfaceAddrs {
		ipNet, ok := address.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ips = append(ips, ipNet.IP)
			}
		}
	}

	return ips
}

// 判断字符串在不在数组中
func InArrayString(s string, arr []string) bool {
	for _, str := range arr {
		if s == str {
			return true
		}
	}
	return false
}