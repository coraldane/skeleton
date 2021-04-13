package skeleton

import (
	"bytes"
	"encoding/json"
	"reflect"

	"github.com/coraldane/logger"

	"github.com/coraldane/skeleton/io"
	"github.com/coraldane/skeleton/proto"
)

func doCheckAuth(session *proto.TcpSession, msg *proto.Message) {
	resp := proto.NewResponse()
	auth := &proto.LoginAuth{}
	err := json.Unmarshal(msg.Body, auth)
	if nil != err {
		logger.Error("unmarshal login auth error", err.Error())
		resp.Code = "common.unmarshal_json_error"
	} else if nil != session.ConnectionListener {
		if authConnListener, ok := session.ConnectionListener.(proto.AuthConnection); ok {
			resp = authConnListener.DoAuth(session, auth)
			if resp.IsSuccess() {
				session.SetUserId(auth.UserId)
			}
		} else {
			resp.Code = "ok"
		}
	} else {
		resp.Code = "ok"
	}

	msg.Body = []byte(resp.String())
	session.Write(msg)
}

func handleRequest(session *proto.TcpSession, msg *proto.Message, tcpRequest *proto.TcpRequest) {
	route, ok := RouteHandler.routes[tcpRequest.Method]
	if !ok {
		session.WriteError(msg, "method not found")
		return
	}

	mt := route.invokeMethod.Type()
	inParams := make([]reflect.Value, mt.NumIn()-1)
	for index := 1; index < mt.NumIn(); index++ {
		inParams[index-1] = reflect.Indirect(reflect.New(mt.In(index)))
	}

	paramBuf := &bytes.Buffer{}
	paramBuf.Write(tcpRequest.Data)
	reader := io.NewReader(paramBuf, true)

	err := reader.ReadArray(inParams)
	if nil != err {
		logger.Error("deserialize params error, method: %s, body: %s, error: %v", tcpRequest.Method, string(tcpRequest.Data), err)
		session.WriteError(msg, "params error")
		return
	}

	inVals := make([]interface{}, len(inParams))
	for index, param := range inParams {
		inVals[index] = param.Interface()
	}

	realParams := make([]reflect.Value, len(inParams)+1)
	ctx := NewContext(session.GetUserId())
	ctx.SessionUniqueKey = session.UniqueKey

	if baseRouter, ok := route.invokeTarget.(BaseRouter); ok {
		baseRouter.InitContext(ctx)
	}

	realParams[0] = reflect.ValueOf(ctx)
	for index, val := range inParams {
		realParams[index+1] = val
	}

	results := route.invokeMethod.Call(realParams)

	if proto.RESPONSE == msg.Cmd { //如果处理的是Response消息，则不再返回
		return
	}

	responseContent := formatResponseContent(results)
	msg.Cmd = uint8(proto.RESPONSE)
	msg.Body = []byte(responseContent)

	logger.Debug("msgId: %s, flag: %d, userId: %d, response: %s", msg.MsgId, msg.Flag, session.GetUserId(), responseContent)
	session.Write(msg)

	ctx.SendPushRequest()
}

func formatResponseContent(results []reflect.Value) string {
	realVals := make([]interface{}, len(results))
	for index, val := range results {
		if resp, ok := val.Interface().(*proto.Response); ok {
			realVals[index] = resp.Value()
		} else {
			realVals[index] = val.Interface()
		}
	}

	bs, _ := json.Marshal(realVals)
	return string(bs)
}
