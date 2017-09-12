package skeleton

import (
	"net"

	"encoding/json"
	"github.com/coraldane/logger"
	"github.com/coraldane/skeleton/proto"
	"gitlab.tarzip.com/udai/HangZhouMaJiang/utils"
)

func ConnectServer(addr string, clientConn proto.ClientConnection) error {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if nil != err {
		return err
	}
	logger.Debug("connect to tcp server", addr)

	session := proto.NewTcpSession(conn, newClientConnListener(clientConn))

	//write ping message first
	msg := proto.NewMessage()
	msg.Cmd = proto.PING
	msg.MsgId = utils.Md5Hex(utils.GetUUID())
	session.Write(msg)

	go handleConnection(session)
	return nil
}

type clientConnListener struct {
	connListener proto.ClientConnection
}

func newClientConnListener(listener proto.ClientConnection) *clientConnListener {
	inst := &clientConnListener{}
	inst.connListener = listener
	return inst
}

func (this *clientConnListener) OnConnected(session *proto.TcpSession) {
	this.connListener.OnConnected(session)
}
func (this *clientConnListener) OnDisconnected(session *proto.TcpSession) {
	this.connListener.OnDisconnected(session)
}

func (this *clientConnListener) HandleMessage(session *proto.TcpSession, msg *proto.Message) {
	this.connListener.HandleMessage(session, msg)

	go func() {
		if err := recover(); nil != err {
			logger.Error("handle message error", err)
		}
	}()

	switch msg.Cmd {
	case proto.PING:
	case proto.ACK:
	case proto.REQUEST:
		tcpRequest := &proto.TcpRequest{}
		tcpRequest.Decode(msg.Body)
		logger.Debug("Request:", msg.MsgId, msg.Flag, tcpRequest.Method, string(tcpRequest.Data))

		if "" == tcpRequest.Method {
			session.WriteError(msg, "method name empty")
			return
		}
		handleRequest(session, msg, tcpRequest)
	case proto.RESPONSE:
		rows := make([]map[string]interface{}, 0)
		err := json.Unmarshal(msg.Body, &rows)
		if nil != err {
			logger.Error("unmarshal data error", msg, err)
		} else if 1 == len(rows) {
			this.connListener.HandleResponse(msg.MsgHead, rows[0])
		} else {
			logger.Error("response data is array", string(msg.Body))
		}
	case proto.PUSH:
		{
			tcpRequest := &proto.TcpRequest{}
			tcpRequest.Decode(msg.Body)

			//返回ACK
			msg.Cmd = proto.ACK
			msg.Body = []byte{}
			session.Write(msg)

			//JSON格式的BODY 换成 序列化格式
			paramMap := make(map[string]interface{})
			err := json.Unmarshal(tcpRequest.Data, &paramMap)
			if nil != err {
				logger.Error("unmarshal data error", msg, err)
			} else {
				this.connListener.HandlePushResponse(msg.MsgHead, tcpRequest.Method, paramMap)
			}
		}
	case proto.ENTER:
		paramMap := make(map[string]interface{})
		err := json.Unmarshal(msg.Body, &paramMap)
		if nil != err {
			logger.Error("unmarshal data error", msg, err)
		} else {
			if "ok" == paramMap["code"] {
				this.connListener.OnAuthSuccess(session)
			}
			this.connListener.HandleResponse(msg.MsgHead, paramMap)
		}
	default:
	}
}
