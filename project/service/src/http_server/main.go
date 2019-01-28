package main

import (
	"log"
	"net/http"
	"fmt"
	"sdk/http_base"
	"sdk/logger"
  _ "net/http/pprof"
)

func BeforServerStart(route_id int) (interface{}, bool) {
	server := &serverLogic{}
	if !server.Initialize() {
		logger.Instance().LogAppError("server logic Initialize Failed!RouteId:%d", route_id)
		return nil, false
	}
    logger.Instance().LogAppInfo("server logic Initialize Ok!RouteId:%d", route_id)

	return server, true
}

func main() {
	err := logger.Instance().Load("../conf/log.xml")
	if err != nil {
		fmt.Printf("log Instance() Failed!ErrString:%s\n", err.Error())
		return
	}
    
	s := http_base.CreateHttpServer()
	s.SetHttpConfig("../conf/server.xml")
	s.SetBeforeInitialFunc(BeforServerStart)
	s.AddHttpFunc("/member/check_login", CheckLogin)
	s.AddHttpFunc("/member/register", Register)
    
    go func() {
        log.Println(http.ListenAndServe(":6060", nil))
    }()

    /*
    go func(){
    time.Sleep(100 * time.Second)
    pprof.StopCPUProfile()
    fmt.Printf("exit")    
    }()
    */
    //go func(){
    if !s.Initialize() {
		logger.Instance().LogAppError("Init http Server Failed!")
	}
    //}()
    
    
}
