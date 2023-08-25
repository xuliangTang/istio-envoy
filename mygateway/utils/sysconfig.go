package utils

import (
	"gopkg.in/yaml.v2"
	"istio-envoy/mygateway/utils/helpers"
	"log"
)

type SysConfigStruct struct {
	NodePort int32 `yaml:"nodeport"`
}

const SysConfigPath = "./sysconfig.yaml"

var SysConfig *SysConfigStruct

func InitSysConfig() {
	SysConfig = &SysConfigStruct{}
	b := helpers.MustLoadFile(SysConfigPath)
	if err := yaml.Unmarshal(b, SysConfig); err != nil {
		log.Fatalln("配置文件错误：", err)
	}
}
