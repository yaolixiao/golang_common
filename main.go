package main

import (
	"github.com/yaolixiao/golang_common/lib"
	_ "net"
	"fmt"
)

func main() {

	// 测试资源初始化
	if err := lib.Init("./conf/dev/"); err != nil {
		fmt.Printf("main init fail. ~~~ err=%v\n", err)
		return
	}

	fmt.Println("init success.")
	fmt.Println("ConfEnvPath=", lib.ConfEnvPath)
	fmt.Println("ConfEnv=", lib.ConfEnv)

	// 测试net相关
	// ips := lib.GetLocationIPs()
	// fmt.Printf("ips=%v\n", ips)
}