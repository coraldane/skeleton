package proto

import (
	"bufio"
	"github.com/coraldane/logger"
	"github.com/coraldane/skeleton/models"
	"github.com/coraldane/utils"
	"net"
	"time"
)

// TcpSession represent a user connection's context
type TcpSession struct {
	Conn               *net.TCPConn // socket
	BR                 *bufio.Reader
	GmtConnect         time.Time
	RemoteAddr         string
	UniqueKey          string
	ConnectionListener Connection
	userId             int64
	hasClosed          bool
}

func NewTcpSession(conn *net.TCPConn, connListener Connection) *TcpSession {
	inst := &TcpSession{}
	inst.Conn = conn
	inst.BR = bufio.NewReader(conn)
	inst.GmtConnect = time.Now()
	inst.RemoteAddr = conn.RemoteAddr().String()
	inst.UniqueKey = utils.GetUUID()

	inst.ConnectionListener = connListener

	if nil != connListener {
		inst.ConnectionListener.OnConnected(inst)
		if _, ok := connListener.(AuthConnection); ok {
			//15秒内如果用户不登陆，则将连接断开
			time.AfterFunc(time.Duration(15)*time.Second, inst.CheckIfLogined)
		}
	}
	return inst
}

func (this *TcpSession) SetUserId(userId int64) {
	this.userId = userId
}

func (this *TcpSession) GetUserId() int64 {
	return this.userId
}

func (this *TcpSession) CheckIfLogined() {
	if this.hasClosed || 0 < this.userId || nil == this.ConnectionListener {
		return
	}
	this.CloseSafely()
}

func (this *TcpSession) IsClosed() bool {
	return this.hasClosed
}

// return false will close the conn
func (this *TcpSession) Read() *Message {
	this.Conn.SetReadDeadline(time.Now().Add(time.Minute))
	defer this.Conn.SetReadDeadline(time.Time{})

	msg := &Message{}
	err := msg.Decode(this.BR)
	if nil != err {
		return nil
	}
	return msg
}

// when write data error, close the connection
func (this *TcpSession) Write(msg *Message) error {
	this.Conn.SetWriteDeadline(time.Now().Add(time.Minute))
	defer this.Conn.SetWriteDeadline(time.Time{})

	msg.Length = len(msg.Body)

	if PING != msg.Cmd {
		logger.Debug("write", this.userId, msg.MsgHead, string(msg.Body))
		tcpMsg := &models.TcpMessage{}
		tcpMsg.MsgId = msg.MsgId
		tcpMsg.Version = msg.Version
		tcpMsg.Cmd = msg.Cmd
		tcpMsg.Flag = msg.Flag
		tcpMsg.Ext = msg.Ext
		tcpMsg.Length = msg.Length
		tcpMsg.Body = string(msg.Body)

		tcpMsg.Save()
	}

	data := msg.Encode()

	var err error
	for i := 0; i < 3; i++ { //最多重试3次
		_, err = this.Conn.Write(data)
		if nil == err {
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}

	logger.Error(err)
	return err
}

func (this *TcpSession) WriteError(msg *Message, body string) error {
	msg.Body = []byte(body)
	msg.Cmd = uint8(ERR)
	return this.Write(msg)
}

func (this *TcpSession) Close() {
	if this.hasClosed {
		return
	}
	this.hasClosed = true

	// net.Conn可以多次关闭
	this.Conn.Close()

	if nil != this.ConnectionListener {
		this.ConnectionListener.OnDisconnected(this)
	}
}

func (this *TcpSession) CloseSafely() {
	if this.hasClosed {
		return
	}

	msg := NewMessage()
	msg.Cmd = CLOSE
	msg.MsgId = utils.Md5Hex(utils.GetUUID())
	this.Write(msg)

	this.hasClosed = true
	if nil != this.ConnectionListener {
		this.ConnectionListener.OnDisconnected(this)
	}

	//3秒后再关闭连接
	time.AfterFunc(time.Second*time.Duration(3), func() {
		// net.Conn可以多次关闭
		this.Conn.Close()
	})
}
