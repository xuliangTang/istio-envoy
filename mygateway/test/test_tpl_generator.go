package main

import (
	"fmt"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"istio-envoy/mygateway/utils"
	"log"
)

const input_tpl_generator = `
{
   "listen_port": 8080
}
`

func main() {
	lis := &listener.Listener{}
	err := utils.NewTplGenerator[*listener.Listener]().
		GetOutput(input_tpl_generator, "listener", lis)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(lis)
}
