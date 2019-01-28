package main

const{
    ERR_OK = 0,
    ERR_INVALID_PARAMS,
    ERR_LOGIN_FAILED,
}

type XmallProtol struct {
	Status     int     `json:"status"`
	Msg        string      `json:"msg"`
	SessionId int64       `json:"sesson_id"`
	Result     interface{} //`json:"result"`
}

func (this *XmallProtol)SetStatus(status int){
    this.Status = status
}

func (this *XmallProtol)SetMsg(msg string){
    this.Msg = msg
}

func (this *XmallProtol)SetSessionId(session_id int64){
    this.SessionId = session_id
}


func (this *XmallProtol)SetResult(result interface{}){
    this.Result = result
}