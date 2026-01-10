package internal

import (
	"fmt"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"sync"
	"yapipt/pkg"

	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)


type ClientConn struct {
	user string
	WSConn *websocket.Conn
	WSConnMutex sync.Mutex
	CloseReaderRoutine bool
}

func loadEnv(ENV_VAR string) (string, error){
	env_var := os.Getenv(ENV_VAR)
	if(env_var==""){
		return "", errors.New(ENV_VAR + " not in env")
	}
	return env_var, nil
}

func (R *Runtime)saveEnv() error {
	var err error

	R.TCPServePort, err = loadEnv("SERVER_TCP_PORT")
	if err != nil {
		return err
	}
	pkg.LogInfo("SERVER_TCP_PORT="+R.TCPServePort)
	return nil
}

type Runtime struct{
	TCPServePort string
	HubMutex sync.Mutex
	WSConnHub map[string]*ClientConn
	BroadcastChan chan []byte

	PSQL_DB *sql.DB
	RedisDB *redis.Client
	DBContext context.Context
}

func (R *Runtime)BroadcastMsgData(raw_bytes []byte) {
	var msg_json pkg.MsgDataJSON
	err := json.Unmarshal(raw_bytes, &msg_json)
	if err!= nil {
		pkg.LogClientError("Unmarshal Error for message_data")
		return 
	}
	for _, CC := range R.WSConnHub {
		CC.WSConnMutex.Lock()
		CC.WSConn.WriteJSON(msg_json)
		CC.WSConnMutex.Unlock()
	}
}

func (R *Runtime)BroadcastMsgIndct(raw_bytes []byte) {
	var msg_json pkg.MsgIndctJSON
	err := json.Unmarshal(raw_bytes, &msg_json)
	if err!= nil {
		pkg.LogClientError("Unmarshal Error for message_data")
		return 
	}
	for _, CC := range R.WSConnHub {
		CC.WSConnMutex.Lock()
		CC.WSConn.WriteJSON(msg_json)
		CC.WSConnMutex.Unlock()
	}
}

func InitRuntime(env_file string) (*Runtime, error) {
	var R Runtime

	if err := R.saveEnv(); err != nil {
		return &R, err
	}

	host := "localhost"
	port := "5432"
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, dbname)
	
	var err error
	R.PSQL_DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return &R, err
	}

	if err = R.PSQL_DB.Ping(); err != nil {
		return &R, fmt.Errorf("ping failed: %v", err)
	}

	R.RedisDB = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	R.WSConnHub = make(map[string]*ClientConn)
	R.DBContext = context.Background()
	R.BroadcastChan = make(chan []byte)

	go func(R *Runtime) {
		for raw_bytes := range R.BroadcastChan {
			if string(raw_bytes) == "" {
				continue
			} else if string(raw_bytes) == "Close" {
				break
			}
			
			var envlp pkg.Envelop
			if err := json.Unmarshal(raw_bytes, &envlp); err != nil {
				continue
			}

			R.HubMutex.Lock()
			switch envlp.Type {
			case pkg.MsgData:
				R.BroadcastMsgData(raw_bytes)
			case pkg.MsgIndct:
				R.BroadcastMsgIndct(raw_bytes)
			}
			R.HubMutex.Unlock()
		}
	}(&R)

	return &R, nil
}

func (R *Runtime) DeInitRuntime() {
	R.BroadcastChan <- []byte("Close")
	for _, CC := range R.WSConnHub {
		CC.WSConnMutex.Lock()
		CC.CloseReaderRoutine = true
		CC.WSConnMutex.Unlock()
	}
}

