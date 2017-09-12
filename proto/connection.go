package proto

type Connection interface {
	OnConnected(session *TcpSession)
	OnDisconnected(session *TcpSession)
	HandleMessage(session *TcpSession, msg *Message)
}

type AuthConnection interface {
	Connection

	DoAuth(session *TcpSession, auth *LoginAuth) *Response
	OnAck(msg *Message)
}

type ClientConnection interface {
	Connection

	OnAuthSuccess(session *TcpSession)
	HandleResponse(msgHead MsgHead, resp map[string]interface{})
	HandlePushResponse(msgHead MsgHead, method string, resp map[string]interface{})
}
