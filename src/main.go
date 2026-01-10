package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"yapipt/internal"
	"yapipt/pkg"
)

func main(){
	R, err := internal.InitRuntime("env")
	if err != nil {
		pkg.LogError("Error Initing the Runtime "+err.Error())
		return
	}
	mux := http.NewServeMux()

	mux.HandleFunc("/ws", R.InitWSConn)
	mux.HandleFunc("/api/user", R.Login)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func(R* internal.Runtime) {
		<- sig
		R.DeInitRuntime()
	}(R)

	pkg.LogInfo("Started at :" + R.TCPServePort)
	err = http.ListenAndServe(":"+R.TCPServePort, mux)
	if err != nil{
		pkg.LogError("http server failed to start at /:" + R.TCPServePort)
	}
}
