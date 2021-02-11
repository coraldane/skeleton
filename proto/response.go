package proto

import (
	"sync"
	"encoding/json"
	"github.com/astaxie/beego/config"
)

var (
	conf config.Configer
)

func init() {
	conf ,_ = config.NewConfig("ini", "i18n.properties")
}

type Response struct {
	sync.RWMutex
	Code    string
	dataMap map[string]interface{}
}

func NewResponse() *Response {
	resp := Response{}
	resp.Code = "unknown"
	resp.dataMap = make(map[string]interface{})
	return &resp
}

func (this *Response) Put(key string, value interface{}) {
	this.Lock()
	defer this.Unlock()

	this.dataMap[key] = value
}

func (this *Response) Get(key string) interface{} {
	this.Lock()
	defer this.Unlock()

	return this.dataMap[key]
}

func (this *Response) IsSuccess() bool {
	return "ok" == this.Code
}

func (this *Response) Build() {
	this.dataMap["code"] = this.Code
	if _, ok := this.dataMap["message"]; !ok {
		this.dataMap["message"] = conf.String(this.Code)
	}
}

func (this *Response) Value() map[string]interface{} {
	this.Build()
	return this.dataMap
}

func (this *Response) String() string {
	this.Build()

	bs, _ := json.Marshal(this.dataMap)
	return string(bs)
	//paramBuf := &bytes.Buffer{}
	//writer := io.NewWriter(paramBuf, true)
	//writer.Serialize(this.dataMap)
	//return paramBuf.String()
}
