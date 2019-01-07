package conn_pool

import (
	"net"
	"sdk/logger"
	"time"
)

type ParseFunc func(recv_buf []byte) ([]byte, int)

type ConnFunc struct {
	Af_conn_func func(net_addr string, has_connected bool) (*[]byte, ParseFunc, interface{})
	Packet_func  func(net_addr string, packet_buf []byte, contxt interface{})
	Af_send_err  func(net_addr string, packet_buf []byte, contxt interface{})
}

type cConn struct {
	conn            net.Conn
	remote_addr     string
	max_packet_size int
	send_chann      chan []byte
	close_chann     chan int
	af_conn_func    func(net_addr string, has_connected bool) (*[]byte, ParseFunc, interface{})
	parse_func      ParseFunc
	packet_func     func(net_addr string, packet_buf []byte, contxt interface{})
	af_send_err     func(net_addr string, packet_buf []byte, contxt interface{})
	call_back_param interface{}
	pool            *cConnPool
	connect_id      int64
}

func (this *cConn) Send(buf []byte) bool {
	time_chann := make(chan int, 0)
	go func() {
		time.Sleep(10 * time.Second)
		time_chann <- 1
	}()

	select {
	case this.send_chann <- buf:
		return true
	case <-time_chann:
		return false
		//case
	}

}

func (this *cConn) Close() {
	this.close_chann <- 1
}

func (this *cConn) RemoteAddr() string {
	return this.remote_addr
}

func (this *cConn) run() {
	log_obj := logger.Instance()
	var send_offset int
	var send_buf []byte
	var recv_offset int
	var recv_buf []byte
	var parse_offset int

	send_offset = 0
	recv_offset = 0
	parse_offset = 0
	recv_buf = make([]byte, 0, this.max_packet_size)

	defer func() {
		this.conn.Close() // 关闭连接
		event := hubEvent{
			event_type: CLOSE_CONNS,
			params:     this,
		}
		this.pool.sendEvent(&event, 10)
	}()

	for {
		// 发送没有发完毕的数据
		if send_buf != nil && len(send_buf) >= send_offset {
			size, err := this.conn.Write(send_buf[send_offset:])
			if err != nil {
				op_err := err.(*net.OpError)
				if op_err.Timeout() { // 发送超时，马上要阻塞
					send_offset += size
					log_obj.LogAccWarn("Need Send Again!RmoteIP=%s,ErrString=%s", this.remote_addr, err.Error())

					break
				} else {
					log_obj.LogAccWarn("Send Err!RmoteIP=%s,ErrString=%s",
						this.remote_addr,
						err.Error())

					return
				}
			}
		}

		send_offset = 0
		if len(this.send_chann) > 0 {
			select {
			case send_buf = <-this.send_chann:
				break
			}
		} else {
			break // 进入到接受数据处理中
		}
	}

	// 尝试读一下是否有关掉连接的信号
	time_chann := make(chan int, 0)
	time_chann <- 1

	select {
	case <-this.close_chann:
		return
	case <-time_chann:
		break
	}

	// 开始接受数据
	for {
		// 定时一秒
		this.conn.SetReadDeadline(time.Now().Add(1 * time.Second))

		size, err := this.conn.Read(recv_buf[recv_offset:])
		if err != nil {
			opt_err := err.(*net.OpError)
			if opt_err.Timeout() {
				log_obj.LogAccInfo("Need Read Again!RemoteIp=%s,ErrString=%s",
					this.remote_addr,
					err.Error())

				break
			} else {
				log_obj.LogSysWarn("Read Failed!RemoteIp=%s,ErrString=%s",
					this.remote_addr,
					err.Error())

				return
			}
		}

		recv_offset += size
		buf, packet_size := this.parse_func(recv_buf[parse_offset:recv_offset])
		if packet_size == 0 { // 解析不完整
			break
		}

		this.packet_func(this.remote_addr, buf, this.call_back_param)
		parse_offset += packet_size

		if recv_offset >= this.max_packet_size {
			copy(recv_buf, recv_buf[parse_offset:recv_offset])
			recv_offset -= parse_offset
			parse_offset = 0
		}

		// 缓冲区已经满
		if recv_offset >= this.max_packet_size {
			log_obj.LogSysError("remoteIP=%s,Packet Full!", this.remote_addr)
			return
		}
	}
}

func (this *cConn) clearSendChann() {
	// 这里使用对send_chann使用range遍历会阻塞，需要调用close函数进行关闭
	close(this.send_chann)
	for v := range this.send_chann {
		go this.af_send_err(this.remote_addr, v, this.call_back_param)
	}
}

func createConn(remote_addr string,
	max_packet_size int,
	set_func ConnFunc,
	pool *cConnPool,
	connect_id int64) *cConn {

	// 连接远程服务
	conn, err := net.DialTimeout("tcp", remote_addr, 10*time.Second)
	if err != nil {
		log_obj := logger.Instance()
		log_obj.LogSysWarn("Connect Failed!RemoteIP=%s,ErrString=%s", remote_addr, err.Error())

		// 连接失败
		_, _, _ = set_func.Af_conn_func(remote_addr, false)
		return nil
	}

	buf, func_, contxt := set_func.Af_conn_func(remote_addr, true)

	conn_client := &cConn{
		send_chann:      make(chan []byte, 1024),
		close_chann:     make(chan int, 0), // 无需等待立即返回
		remote_addr:     remote_addr,
		af_conn_func:    set_func.Af_conn_func,
		parse_func:      func_,
		af_send_err:     set_func.Af_send_err,
		conn:            conn,
		pool:            pool,
		max_packet_size: max_packet_size,
		call_back_param: contxt,
		packet_func:     set_func.Packet_func,
		connect_id:      connect_id,
	}

	go conn_client.run()

	return conn_client
}
