package event_route

import (
	"sdk/logger"
	"time"
)

type EventMsg struct {
	Handle       func(route_id int, context interface{}, init_param interface{})
	Context      interface{}
	Wait_channel chan int
	Elapsed_time time.Duration // 投递到消息队列中的时间戳
}

type EventRouteMgr struct {
	max_event_nums  int
	route_channel   []chan *EventMsg
	params_array    []interface{}
	route_nums      int
	pending_time    int // 单位:ms
	log             logger.ILogger
	snow_slide_time int // 防雪崩时间 单位:ms
	//alaram_func    func(rout_id int, init_param interface{})
}

func CreateEventRouteMgr(max_event_nums int,
                    	route_nums int,
                    	pending_time int,
                    	snow_slide_time int,
                    	init_func func(route_id int) (interface{}, bool)) *EventRouteMgr {

	event_mgr := &EventRouteMgr{
		max_event_nums:  max_event_nums,
		pending_time:    pending_time,
		route_nums:      route_nums,
		route_channel:   make([]chan *EventMsg, route_nums),
		log:             logger.Instance(),
		params_array:    make([]interface{}, route_nums),
		snow_slide_time: snow_slide_time,
	}

	// 初始化携程
	for i := 0; i < route_nums; i ++ {
		init_param, ok := init_func(i)
		if !ok {
			event_mgr.log.LogSysError("Params Initialize()Failed!RouteId:%d", i)
			return nil
		}
		event_mgr.params_array[i] = init_param
	}

	for i := 0; i < route_nums; i ++ {
		event_mgr.route_channel[i] = make(chan *EventMsg, max_event_nums)
	}

	// 开启协程
	for i := 0; i < route_nums; i++ {
		go event_mgr.run(i)
	}

	return event_mgr
}

func (this *EventRouteMgr) SyncHandle(route_id int, event *EventMsg, wait_time int) int {
	index := route_id % this.route_nums
	time_channel := make(chan int)
	go func() {
		time.Sleep(time.Duration(this.pending_time) * time.Millisecond)
		time_channel <- 1
	}()

	select {
		case this.route_channel[index] <- event:
		case <-time_channel:
			channel_cap := len(this.route_channel[index])
			this.log.LogSysError("route:%d is full!nums:%d", index, channel_cap)
			return -1
	}
    
    
    // 等待处理完毕
	go func() {
		time.Sleep(time.Duration(wait_time) * time.Millisecond)
        this.log.LogSysError("Waiting TimeOut,dispatchId:%d", index)
		time_channel <- 1
	}()

	// 等待应带处理完毕
	select {
		case <-event.Wait_channel:
			return 1
		case <-time_channel:
			return 0
	}
}

func (this *EventRouteMgr) run(route_id int) {
	channel := this.route_channel[route_id]
	init_param := this.params_array[route_id]
	for {
		select {
		case event := <-channel: // 收到消息进行处理
			// 防雪崩操作
            now_micro := time.Duration(time.Now().UnixNano())
			after_micro := now_micro - event.Elapsed_time


			// 前面处理的时间太长了，此数据包丢弃掉
			if after_micro >= time.Duration(this.snow_slide_time) * time.Millisecond {
                
				// 记录日志
				this.log.LogSysError("snow-slide happened!ElaspedTime:%dms,Now:%dms,keep-time:%dms,TaskNums:%d,route_id:%d",
                                     event.Elapsed_time / time.Millisecond,
                                     now_micro / time.Millisecond,
                                     after_micro / time.Millisecond,
                                     len(channel),
                                     route_id)
				continue
			}

			event.Handle(route_id, event.Context, init_param)
			event.Wait_channel <- 1

		}
	}
}
