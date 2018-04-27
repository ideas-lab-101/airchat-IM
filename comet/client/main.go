package main

import (
	"flag"
	"runtime"

	log "github.com/thinkboy/log4go"

	//***abc自己的代码
	"fmt"
)

func main() {
	flag.Parse()
	if err := InitConfig(); err != nil {
		panic(err)
	}
	runtime.GOMAXPROCS(Conf.MaxProc)
	log.LoadConfiguration(Conf.Log)
	defer log.Close()

	// fmt.Println(">start link ", Conf)

	if Conf.Type == ProtoTCP {
		fmt.Println(">initTCP")
		initTCP()
	} else if Conf.Type == ProtoWebsocket {
		fmt.Println(">initWebsocket")
		initWebsocket()
	} else if Conf.Type == ProtoWebsocketTLS {
		fmt.Println(">initWebsocketTLS")
		initWebsocketTLS()
	}
}
