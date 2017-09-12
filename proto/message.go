package proto

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/coraldane/logger"
)

const (
	BYTE_PREFIX = 0xfa
)

const (
	PING uint8 = iota + 1 // 1
	ENTER
	REQUEST
	RESPONSE
	PUSH
	ACK
	CLOSE
	ERR = 0xff
)

/**
* TCP包头信息
* 头格式：23字节(固定值0xfa[1字节] + 版本号[1字节] + 指令[1字节] + Flag[1字节] + 预留位[1字节] + length[2字节] + msgId[16字节])
* 版本号: 0x01 - JSON序列化
* 指令
  0x01 请求消息
  0x02 ACK消息
* Flag:  留给客户端标记的字节
* 预留位: 1字节
* length: Body的长度，不包括包头长度
* msgId: 消息ID[MD5值, 如 0830b69ce8af7a5e6b21658bbf5f0eeb]转换规则为每两位作为一个字节,0x08,0x30
*/
type MsgHead struct {
	Version uint8
	Cmd     uint8
	Flag    uint8
	Ext     uint8
	Length  int
	MsgId   string
}

type Message struct {
	MsgHead
	Body []byte
}

func NewMessage() *Message {
	inst := &Message{}
	inst.Version = 1
	return inst
}

func (this *Message) Decode(br *bufio.Reader) error {
	defer func() {
		if err := recover(); nil != err {
			logger.Error("Message Decode recover err", err)
		}
	}()

	headBytes, err := readBytes(br, 22)
	if nil != err {
		return err
	}

	this.Version = uint8(headBytes[0])
	this.Cmd = uint8(headBytes[1])
	this.Flag = uint8(headBytes[2])
	this.Ext = uint8(headBytes[3])
	this.Length = int(binary.BigEndian.Uint16(headBytes[4:6]))

	msgIdBytes := headBytes[6:]
	var msgId bytes.Buffer
	for _, bt := range msgIdBytes {
		msgId.WriteString(fmt.Sprintf("%02x", bt))
	}
	this.MsgId = msgId.String()

	data, err := readBytes(br, this.Length)
	if nil != err {
		return err
	}
	this.Body = data

	return nil
}

func (this *Message) Encode() []byte {
	var data bytes.Buffer
	data.WriteByte(BYTE_PREFIX)
	data.WriteByte(this.Version)
	data.WriteByte(this.Cmd)
	data.WriteByte(this.Flag)
	data.WriteByte(this.Ext)

	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(this.Length))
	data.Write(buf)

	for index := 0; index < len(this.MsgId); index = index + 2 {
		val, _ := strconv.ParseInt(this.MsgId[index:index+2], 16, 0)
		data.WriteByte(byte(val))
	}

	data.Write(this.Body)
	return data.Bytes()
}

func readBytes(br *bufio.Reader, length int) ([]byte, error) {
	data := make([]byte, length)
	for m := 0; m < len(data); {
		n, err := br.Read(data[m:])
		if err != nil || n <= 0 {
			return nil, err
		}
		m += n
	}
	return data, nil
}
