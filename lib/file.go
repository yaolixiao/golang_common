package lib

import (
	"strings"
	"fmt"
	"os"
	"io/ioutil"
	"bytes"
	"github.com/spf13/viper"
)

var ConfEnvPath string  // 配置文件夹
var ConfEnv string 		// 配置环境名 比如：dev prod test

// 解析配置文件目录: 将配置文件夹和配置环境赋给全局变量
// 
// 配置文件必须放到文件夹内
// 如：配置文件是conf/dev/base.json 则 ConfEnvPath=conf/dev	 ConfEnv=dev 
func ParseConfPath(confPath string) error {
	paths := strings.Split(confPath, "/")
	folder := strings.Join(paths[:len(paths)-1], "/")
	ConfEnvPath = folder
	ConfEnv = paths[len(paths)-2]
	return nil
}

func GetConfPath(fileName string) string {
	return ConfEnvPath + "/" + fileName + ".toml"
}

func ParseConfig(path string, conf interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open config path [%v] fail. err=%v", path, err)
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("read config path [%v] fail. err=%v", path, err)
	}

	v := viper.New()
	v.SetConfigType("toml")
	v.ReadConfig(bytes.NewBuffer(data))
	if err := v.Unmarshal(conf); err != nil {
		return fmt.Errorf("Parse config fail. config=%v, err=%v", string(data), err)
	}
	return nil
}