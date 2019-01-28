package http_base

import (
	"html/template"
	"strconv"
	"net/http"
	"sdk/logger"
	"time"
)

type HttpRequest http.Request

func (this HttpRequest) GetHeader(key string) string {
	return this.Header.Get(key)
}

func (this HttpRequest) GetCookie(key string) string {
	request := (http.Request)(this)
	cookie, err := request.Cookie(key)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (this HttpRequest) GetParam(key string) string {
	query := this.URL.Query()
	value := query.Get(key)
    value = template.JSEscapeString(value)
	return value
}

func (this HttpRequest)GetParamI(key string) int64 {
    query := this.URL.Query()
    value := query.Get(key)
    v,err := strconv.ParseInt(value, 10, 64)
    if err != nil {
        v = 0    
    }
    
    return v
}

func (this HttpRequest) GetPost(data []byte) int {
	size, _ := this.Body.Read(data)
	return size
}

type HttpResponse struct {
	reponse_header map[string]string
	cookies        []*http.Cookie
	data           []byte
	localtion      string // 302重定向
}

func (this *HttpResponse) SetHeader(key string, value string){
    if this.reponse_header == nil{
        this.reponse_header = make(map[string]string, 10)
    }
    this.reponse_header[key] = value
}
func (this *HttpResponse) SetLocation(localtion string) {
	localtion = localtion
}

func (this *HttpResponse) setResponseWriter(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	for k, v := range this.reponse_header {
		w.Header().Set(k, v)
	}
	for _, v := range this.cookies {
		http.SetCookie(w, v)
	}

	if this.localtion != "" {
		w.WriteHeader(http.StatusFound)
	} else {
		_, err := w.Write(this.data)
		if err != nil {
			logger.Instance().LogAccError("Response Data Failed!ErrString:%s", err.Error())
		}
	}
}

func (this *HttpResponse) SetCookie(key string, value string, domain string, path string, expire_time int) {
	cookie := &http.Cookie{
		Name:     key,
		Value:    value,
		Domain:   domain,
		Path:     path,
		Expires:  time.Now().Add(time.Duration(expire_time) * time.Second),
		HttpOnly: true,
	}
    if this.cookies == nil{
        this.cookies = make([]*http.Cookie, 0, 10)
    }
    this.cookies = append(this.cookies, cookie)
}

func (this *HttpResponse)SetData(data []byte){
    this.data = data
}
