package skeleton

import (
	"github.com/coraldane/logger"
	"github.com/coraldane/skeleton/proto"
	"net"
	"time"
)

func StartServer(addr string, connListener proto.Connection, tcpServer *TcpServer) error {
	tcpAddr, _ := net.ResolveTCPAddr("tcp", addr)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if nil != err {
		return err
	}
	defer listener.Close()

	logger.Info("listen tcp server at", addr)

	if nil == tcpServer {
		tcpServer = NewTcpServer()
	}

	for {
		conn, err := listener.AcceptTCP()
		if nil != err {
			logger.Error("accept new client error", err)
			continue
		}

		conn.SetKeepAlive(true)
		conn.SetKeepAlivePeriod(time.Second)

		session := proto.NewTcpSession(conn, newServerConnection(connListener, tcpServer))
		go handleConnection(session)
	}
}

func handleConnection(session *proto.TcpSession) {
	defer session.Conn.Close()

	for {
		sbyte, err := session.BR.ReadByte()
		if nil == err {
			if proto.BYTE_PREFIX != sbyte {
				continue
			}
		} else {
			logger.Error("read data from conn error", err)
			session.Close()
			break
		}

		msg := session.Read()
		if nil == msg {
			continue
		}

		session.ConnectionListener.HandleMessage(session, msg)
	}
}

type serverConnection struct {
	connListener proto.Connection
	tcpServer    *TcpServer
}

func newServerConnection(listener proto.Connection, tcpServer *TcpServer) *serverConnection {
	inst := &serverConnection{}
	inst.connListener = listener
	inst.tcpServer = tcpServer
	return inst
}

func (this *serverConnection) OnConnected(session *proto.TcpSession) {
	this.tcpServer.PutSession(session.UniqueKey, session)
	logger.Info("accept conn from ", session.RemoteAddr, ", conn num:", this.tcpServer.SessionCount())

	this.connListener.OnConnected(session)
}

func (this *serverConnection) OnDisconnected(session *proto.TcpSession) {
	this.tcpServer.DeleteSession(session.UniqueKey)
	logger.Info("connection closed ", session.RemoteAddr, session.GetUserId(), ", conn num:", this.tcpServer.SessionCount())

	this.connListener.OnDisconnected(session)
}

func (this *serverConnection) OnAck(msg *proto.Message) {
	if authConnListener, ok := this.connListener.(proto.AuthConnection); ok {
		authConnListener.OnAck(msg)
	}
}

func (this *serverConnection) HandleMessage(session *proto.TcpSession, msg *proto.Message) {
	this.connListener.HandleMessage(session, msg)

	go func() {
		if err := recover(); nil != err {
			logger.Error("handle message error", err)
		}
	}()

	switch msg.Cmd {
	case proto.PING:
		session.Write(msg)
	case proto.ENTER:
		logger.Debug("Enter:", string(msg.Body))
		doCheckAuth(session, msg)
	case proto.ACK:
		this.OnAck(msg)
	case proto.REQUEST:
		tcpRequest := &proto.TcpRequest{}
		tcpRequest.Decode(msg.Body)
		logger.Debug("Request:", msg.MsgId, msg.Flag, tcpRequest.Method, string(tcpRequest.Data))

		if "" == tcpRequest.Method {
			session.WriteError(msg, "method name empty")
			return
		}
		handleRequest(session, msg, tcpRequest)
	default:
	}
}

func (this *serverConnection) DoAuth(session *proto.TcpSession, auth *proto.LoginAuth) *proto.Response {
	resp := proto.NewResponse()
	if authConnListener, ok := this.connListener.(proto.AuthConnection); ok {
		resp = authConnListener.DoAuth(session, auth)
	} else {
		resp.Code = "ok"
	}
	return resp
}
