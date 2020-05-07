package main

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

type LogCfg struct {
	LogLevel    string `yaml:"level"`       //日志等级
	MaxFileSize int64  `yaml:"filesize"`    //最大日志文件大小（M）
	BackendName string `yaml:"backendname"` //后端名(rpc)
	ServerName  string `yaml:"servername"`  //服务名(service)
	LogField    string `yaml:"logfield"`    //日志打印域控制
}

func LoadYamlConfig() (*LogCfg,error){
	var cfgPath  = DefaultPath + "\\src\\" + "logrus-extends\\config.yaml"
	if !filepath.IsAbs(cfgPath) {
		file, err := filepath.Abs(cfgPath)
		if err != nil {
			return nil,err
		}
		cfgPath = file
		
	}
	data, err := ioutil.ReadFile(cfgPath)
	if err != nil{
		return nil,err
	}
	
	cfg := new(LogCfg)
	if err = yaml.Unmarshal(data, cfg); err != nil{
		logrus.Error("yaml.Unmarshal.err:",err)
		return nil,err
	}
	return cfg,nil
}



