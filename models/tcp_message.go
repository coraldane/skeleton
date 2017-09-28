package models

import (
	"github.com/astaxie/beego/orm"
	"time"
)

type TcpMessage struct {
	Id        int64 `orm:"auto"`
	GmtCreate time.Time
	MsgId     string
	Version   uint8
	Cmd       uint8
	Flag      uint8
	Ext       uint8
	Length    int
	Body      string
}

func (this *TcpMessage) TableUnique() [][]string {
	return [][]string{
		[]string{"MsgId"},
	}
}

func (this *TcpMessage) Save() (int64, error) {
	this.GmtCreate = time.Now()
	return orm.NewOrm().Insert(this)
}
