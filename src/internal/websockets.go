package internal

import (
	"encoding/json"
	"net/http"
	"yapipt/pkg"

	"github.com/gorilla/websocket"
)

func (R *Runtime)InitWSConn(w http.ResponseWriter, r *http.Request) {
	WSProtoUpgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
				return true 
		},
	}

	conn, err := WSProtoUpgrader.Upgrade(w,r,nil);
	if err != nil{
		pkg.LogClientError("WS proto upgrade failed for client " + r.RemoteAddr)
		return
	}
	
	user := r.URL.Query().Get("user")
	if(user==""){
		pkg.LogClientError("Invalid params for WS proto upgrade for client " + r.RemoteAddr)
		conn.WriteMessage(websocket.TextMessage , []byte("Invalid URL Parameters"))
		conn.Close()
		return
	}

	CC := &ClientConn{ user: user, WSConn: conn, CloseReaderRoutine: false}

	go func(CC *ClientConn, R *Runtime) {
		conn = CC.WSConn
		var rawBytes []byte
		var err error
		for(!CC.CloseReaderRoutine){
			_, rawBytes, err = conn.ReadMessage()
			if(err != nil) {
				R.HubMutex.Lock()
				delete(R.WSConnHub, CC.user)
				R.HubMutex.Unlock()
				CC.CloseReaderRoutine = true
				break
			}
			var envlp pkg.Envelop
			err = json.Unmarshal(rawBytes, &envlp)
			if err!=nil{
				pkg.LogClientError("Unmarshal Error for rawBytes from client")
			}
			if(envlp.Type==pkg.MsgData){
				R.BroadcastChan <- rawBytes
			}
			CC.WSConnMutex.Lock()
			CC.WSConn.WriteMessage(websocket.TextMessage, []byte("Received"))
			CC.WSConnMutex.Unlock()
		}
	}(CC, R)

	R.HubMutex.Lock()
	R.WSConnHub[user] = CC
	R.HubMutex.Unlock()

	conn.WriteMessage(websocket.TextMessage , []byte("WS Connection Extablished"))
}
