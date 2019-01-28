package main

import (
	"encoding/json"
	"fmt"
	"sdk/http_base"
	"sdk/logger"
)

type serverLogic struct {
}

func (this *serverLogic) Initialize() bool {
	return true
}

//func(w *HttpResponse, r *HttpRequest, param interface{})

// 输出
func CheckLogin(w *http_base.HttpResponse, r *http_base.HttpRequest, param interface{}) {
	callback := r.GetParam("callback")
    login_name := r.GetParam("login_name")
    passwd := r.GetParam("passwd")
    login_type := r.GetParam("type")
    encrpt_method := r.GetParamI("encrpt_method")

	var response_data XmallProtol
	response_data.SetStatus(0)
	response_data.SetSessionId(0)
	response_data.SetMsg("ok")

	if test == "1" {
		data := "hello"
		response_data.SetResult(data)
	} else {
		response_data.SetResult("world")
	}

	data_res := jsonP(callback, &response_data)
    logger.Instance().LogAppDebug("Response:%s", data_res)
	if data_res != nil {
		w.SetData(data_res)
	}
}

func Register(w *http_base.HttpResponse, r *http_base.HttpRequest, param interface{}) {

}

func jsonP(callback string, protol *XmallProtol) []byte {
	data_res, err := json.Marshal(protol)
	if err != nil {
		logger.Instance().LogAppError("Json Marshal Failed!ErrString:%s", err.Error())
		return nil
	}

	jsonp := fmt.Sprintf("try %s(%s) catch(e){}", callback, data_res)
	return []byte(jsonp)
}
