package skeleton

import (
	"github.com/coraldane/logger"
	"github.com/coraldane/skeleton/models"
	"github.com/coraldane/toolkits/container"
	"time"
)

type Context struct {
	userId           int64
	SessionUniqueKey string
	pushFunc         func(*models.PushRequest)
	pushRequestList  *container.SafeList[*pushMessage]
}

func NewContext(otherId int64) *Context {
	inst := &Context{}
	inst.userId = otherId
	inst.pushRequestList = container.NewSafeList[*pushMessage]()
	return inst
}

func (this *Context) SetUserId(otherId int64) {
	this.userId = otherId
}

func (this *Context) GetUserId() int64 {
	return this.userId
}

func (this *Context) SetPushFunc(pushFunc func(*models.PushRequest)) {
	this.pushFunc = pushFunc
}

func (this *Context) AddPushRequest(request *models.PushRequest) {
	this.pushRequestList.PushFront(&pushMessage{request, 0})
}

func (this *Context) AddDelayRequest(request *models.PushRequest, delay time.Duration) {
	this.pushRequestList.PushFront(&pushMessage{request, delay})
}

func (this *Context) SendPushRequest() {
	requestSize := this.pushRequestList.Len()
	if 0 == requestSize {
		return
	}

	if nil == this.pushFunc {
		logger.Error("push request func not set.")
		return
	}

	for 0 < this.pushRequestList.Len() {
		pushMsg := this.pushRequestList.PopBack()
		if 0 < pushMsg.delay {
			time.Sleep(pushMsg.delay)
		}
		this.pushFunc(pushMsg.request)
	}
}

type pushMessage struct {
	request *models.PushRequest
	delay   time.Duration
}
