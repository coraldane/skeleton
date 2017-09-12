package proto

import (
	"bytes"
	"encoding/binary"

	"github.com/coraldane/skeleton/io"
)

/**
 * method length[1字节] + method name + params length[2字节] + params
 */
type TcpRequest struct {
	Method string
	Data   []byte
}

func (this *TcpRequest) Encode(paramData string) []byte {
	var data bytes.Buffer

	data.WriteByte(byte(len(this.Method)))
	data.WriteString(this.Method)

	paramsLen := len(paramData)
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(paramsLen))
	data.Write(buf)
	data.WriteString(paramData)

	return data.Bytes()
}

func (this *TcpRequest) Decode(data []byte) {
	var pos int
	if nil == data || 4 > len(data) {
		return
	}

	methodLen := data[0]
	pos = int(methodLen) + 1
	if pos > len(data) {
		return
	}
	this.Method = string(data[1:pos])
	if (pos + 2) > len(data) {
		return
	}
	paramsLen := int(binary.BigEndian.Uint16(data[pos : pos+2]))

	if (pos + 2 + paramsLen) > len(data) {
		return
	}
	this.Data = data[pos+2 : pos+paramsLen+2]
}

func EncodeTcpRequest(method string, params []interface{}) []byte {
	request := &TcpRequest{}
	request.Method = method

	paramBuf := &bytes.Buffer{}
	writer := io.NewWriter(paramBuf, true)
	writer.Serialize(params)
	return request.Encode(paramBuf.String())
}
