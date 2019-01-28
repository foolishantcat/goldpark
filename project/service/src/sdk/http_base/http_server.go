package http_base

import (
	"fmt"
	"net/http"
	"sdk/event_route"
	"sdk/logger"
	"time"
)

type HttpFunc func(w *HttpResponse, r *HttpRequest, param interface{})

func setDefHttpResponse(response http.ResponseWriter){
	//response.WriteHeader(http.StatusOK) // 默认正常
	//response.Header().Set("Content-Type", "text/html;charset=utf-8")
	response.Header().Set("Transfer-Encoding", "chunked")
	response.Header().Set("Cache-Control", "no-cache")
	response.Header().Set("Server", "go-http-server")
	response.Header().Set("Date", time.Now().String())
	response.Header().Set("Expires", time.Now().String())
    response.Header().Set("Connection",  "keep-alive")
}

type httpData struct {
    reqeust  *HttpRequest
	response *HttpResponse
	handle   HttpFunc
}
type IHttpServer interface {
	SetHttpConfig(conf_path string) bool
	SetBeforeInitialFunc(handle func(route_id int) (interface{}, bool))
	AddHttpFunc(func_name string, handle HttpFunc)
	Initialize() bool
}

func CreateHttpServer() IHttpServer {
	return &httpServer{
        func_map: make(map[string]HttpFunc, 10),
        bind_port: 8080,
        wait_for_handle_time: 200,     // 等待消息处理时间200ms
        opt_time_out: 3000,
        log: logger.Instance(),
    }
}

type httpServer struct {
	func_map     map[string]HttpFunc
	bind_address string                     // 绑定IP
	bind_port    int                        // 绑定端口
    wait_for_handle_time int                // 投递协程池消息等待时间，单位:ms
    opt_time_out    int                        // 收包等待时间 单位:ms
	log          logger.ILogger
	event_mgr    *event_route.EventRouteMgr
	before_func  func(route_id int) (interface{}, bool)
}

func (this *httpServer) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	url_name := request.URL.Path
	hash_id := int(time.Duration(time.Now().UnixNano()) / time.Millisecond) // 用时间戳作为hash_key(单位：ms)
	
	route_func, ok := this.func_map[url_name]
    setDefHttpResponse(response)
	if !ok {
        
        // ResponseWriter接口设置返回的消息类型时候，只能够设置一次，后面设置多次无效
        response.WriteHeader(http.StatusNotFound)
		// 没有找到路由,直接返回404
		this.log.LogSysInfo("Disptach Not Find!url:%s,IP:%s", url_name, request.RemoteAddr)
        
        return
	}
    
    http_response := &HttpResponse{}
	http_data := &httpData{
		response: http_response,
		reqeust:  (*HttpRequest)(request),
		handle:   route_func,
	}
	event := event_route.EventMsg{
		Handle:       httpHandle,
		Context:      http_data,
		Wait_channel: make(chan int, 100),
        Elapsed_time: time.Duration(time.Now().UnixNano()),
	}

	ret:=this.event_mgr.SyncHandle(hash_id, &event, this.wait_for_handle_time)
	if ret <= 0 {
        
		// 处理超时,返回http处理超时
		this.log.LogSysError("SyncHandle TimeOut!Ret:%d,UrlName%s", ret, url_name)
		response.WriteHeader(http.StatusInternalServerError)
            
        return
    }
    http_response.setResponseWriter(response)
    
}

func (this *httpServer) SetHttpConfig(conf_path string) bool {
	return true
}

func (this *httpServer) SetBeforeInitialFunc(handle func(route_id int) (interface{}, bool)) {
	this.before_func = handle
}

func (this *httpServer) AddHttpFunc(func_name string, handle HttpFunc) {
		this.func_map[func_name] = handle
}

func httpHandle(route_id int, context interface{}, init_param interface{}) {
    
   // 回调给上层进行处理
    handle_data := context.(*httpData)
    handle_data.handle(handle_data.response, handle_data.reqeust, init_param)
}

func (this *httpServer) Initialize() bool {
	// 加载conf文件配置绑定端口和信息
	address := fmt.Sprintf("%s:%d", this.bind_address, this.bind_port)
	log := logger.Instance()

	max_packet_nums := 50
	max_route_nums := 100
	pending_time := 200                 	// 投递消息等待200ms
	snow_slide_time := 1000             	// 雪崩处理时间1000ms
   	this.wait_for_handle_time = 500    	// 等待消息处理时间500ms
    this.opt_time_out = 3000                  // 收包超时3000ms
	this.event_mgr = event_route.CreateEventRouteMgr(max_packet_nums,
                                                    max_route_nums,
										            pending_time,
										            snow_slide_time,
										            this.before_func)

	if this.event_mgr == nil {
		return false
	}

	// 创建httplistener
    // 通过serve对象才能够设置超时,ListenAndServe()函数无法设置超时
    server:= &http.Server{
        Addr: address,
        Handler: this,
        ReadTimeout: time.Duration(this.opt_time_out) * time.Millisecond,
        WriteTimeout: time.Duration(this.opt_time_out) * time.Millisecond,
        
    }
    err := server.ListenAndServe()
    //err := http.ListenAndServe(address, this)
	if err != nil {
		log.LogSysError("ListenAndServe Failed!Address:%s,ErrString:%s", address, err.Error())
		return false
	}

	return true
}
