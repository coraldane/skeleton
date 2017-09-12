package models

import (
	"encoding/json"
	"time"
)

type PushRequest struct {
	Id         int64
	MsgId      string
	GmtCreate  time.Time
	RoomId     int64
	UserId     int64
	MethodName string
	Data       string `orm:"type(text)"`
}

func (this *PushRequest) TableIndex() [][]string {
	return [][]string{
		[]string{"UserId"}, []string{"MsgId"},
	}
}

func (this *PushRequest) SetData(v interface{}) {
	bs, err := json.Marshal(v)
	if nil == err {
		this.Data = string(bs)
	}
}
