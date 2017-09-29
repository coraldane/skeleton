package skeleton

import (
	"github.com/coraldane/skeleton/proto"
	"sync"
)

type TcpServer struct {
	m_sessionDict *sync.Map
}

func NewTcpServer() *TcpServer {
	inst := TcpServer{}
	inst.m_sessionDict = &sync.Map{}
	return &inst
}

func (this *TcpServer) PutSession(uniqueKey string, session *proto.TcpSession) {
	this.m_sessionDict.Store(uniqueKey, session)
}

func (this *TcpServer) GetSession(uniqueKey string) *proto.TcpSession {
	if val, ok := this.m_sessionDict.Load(uniqueKey); ok {
		return val.(*proto.TcpSession)
	}
	return nil
}

func (this *TcpServer) DeleteSession(uniqueKey string) {
	this.m_sessionDict.Delete(uniqueKey)
}

func (this *TcpServer) GetUniqueKeys() []string {
	keys := make([]string, 0)
	this.m_sessionDict.Range(func(key, val interface{}) bool {
		if text, ok := key.(string); ok {
			keys = append(keys, text)
		}
		return true
	})
	return keys
}

func (this *TcpServer) SessionCount() int {
	retValue := 0
	this.m_sessionDict.Range(func(key, val interface{}) bool {
		retValue += 1
		return true
	})
	return retValue
}
