package internal

import (
	"encoding/json"
	"net/http"
	"strings"
	"yapipt/pkg"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

func (R *Runtime)InitWSConn(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")
	session_token_cookie, err := r.Cookie("session_token")
	session_token_from_redis, err := R.RedisDB.Get(R.DBContext, user).Result()
	if err != nil {
		if err != redis.Nil{
			w.WriteHeader(http.StatusInternalServerError)
			pkg.LogError("login state improper " + err.Error())
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		pkg.LogError("Error Querying RedisDB " + err.Error())
		return
	}
	session_token := strings.Split(session_token_cookie.String(), "=")
	if session_token_from_redis != session_token[1] {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

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
			R.BroadcastChan <- rawBytes	
			CC.WSConnMutex.Lock()
			CC.WSConn.WriteMessage(websocket.TextMessage, []byte("Received"))
			CC.WSConnMutex.Unlock()
		}
		left_client := pkg.MsgIndctJSON{Type: pkg.MsgIndct, IndctType: pkg.Left, User: CC.user}
		send_bytes, err := json.Marshal(left_client)
		if err != nil {
			pkg.LogWarn("Failed to Marshal joined_client")
		}
		R.BroadcastChan <- send_bytes
	}(CC, R)

	R.HubMutex.Lock()
	R.WSConnHub[user] = CC
	R.HubMutex.Unlock()

	CC.WSConnMutex.Lock()
	CC.WSConn.WriteMessage(websocket.TextMessage , []byte("WS Connection Extablished"))
	CC.WSConnMutex.Unlock()
	
	joined_client := pkg.MsgIndctJSON{Type: pkg.MsgIndct, IndctType: pkg.Joined, User: user}
	send_bytes, err := json.Marshal(joined_client)
	if err != nil {
		pkg.LogWarn("Failed to Marshal joined_client")
	}
	R.BroadcastChan <- send_bytes
}
