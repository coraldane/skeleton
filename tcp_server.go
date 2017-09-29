package skeleton

import (
	"github.com/coraldane/skeleton/proto"
	"sync"
)

type TcpServer struct {
	sync.RWMutex
	m_sessionDict map[string]*proto.TcpSession
}

func NewTcpServer() *TcpServer {
	inst := TcpServer{}
	inst.m_sessionDict = make(map[string]*proto.TcpSession)
	return &inst
}

func (this *TcpServer) PutSession(uniqueKey string, session *proto.TcpSession) {
	this.Lock()
	defer this.Unlock()

	this.m_sessionDict[uniqueKey] = session
}

func (this *TcpServer) GetSession(uniqueKey string) *proto.TcpSession {
	this.Lock()
	defer this.Unlock()
	return this.m_sessionDict[uniqueKey]
}

func (this *TcpServer) DeleteSession(uniqueKey string) {
	this.Lock()
	defer this.Unlock()
	delete(this.m_sessionDict, uniqueKey)
}

func (this *TcpServer) GetUniqueKeys() []string {
	this.Lock()
	defer this.Unlock()

	keys := make([]string, 0)
	for kv, _ := range this.m_sessionDict {
		keys = append(keys, kv)
	}
	return keys
}

func (this *TcpServer) SessionCount() int {
	return len(this.m_sessionDict)
}
