package main

import (
	"github.com/yaolixiao/golang_common/lib"
	_ "net"
	"fmt"
)

func main() {

	// 测试资源初始化
	if err := lib.Init("./conf/dev/"); err != nil {
		fmt.Printf("main init fail. err=%v\n", err)
		return
	}

	fmt.Println("init success.")
	// fmt.Println("ConfEnvPath=", lib.ConfEnvPath)
	// fmt.Println("addr=", lib.GetStringConf("base.http.addr"))
	// fmt.Println("max_header_bytes=", lib.GetIntConf("base.http.max_header_bytes"))
	// fmt.Println("write_timeout=", lib.GetIntConf("base.http.write_timeout"))
	// fmt.Println("read_timeout=", lib.GetIntConf("base.http.read_timeout"))
	// fmt.Println("lib.ConfBase.DebugMode=", lib.ConfBase.DebugMode)


	// 测试net相关
	// ips := lib.GetLocationIPs()
	// fmt.Printf("ips=%v\n", ips)
}