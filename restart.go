package server

import (
	"fmt"
	"golang.org/x/net/websocket"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var origin = "http://127.0.0.1:10002/"
var url = "ws://127.0.0.1:10002/Process"

func (this *XServer) ConnectToProcess() {
	fmt.Println("os.Args[0] : ", os.Args[0])
	filePath := getCurrentDirectory()
	//文件名
	fileName := filepath.Base(os.Args[0])

	ws := connect()

	if ws != nil {
		fmt.Println("ws is not nil")
		ws.Write([]byte(`{"path":"` + fileName + `","fullPath":"` + filePath + `"}`))
		buf := make([]byte, 256)

		_, err := ws.Read(buf)

		if err != nil {
			fmt.Println(err)
		}

		heart(ws)
	} else {
		fmt.Println("ws is nil")
	}


	//path, _ := filepath.Abs(filePath)

	//fmt.Println("filePath:",filePath," fileName:",fileName)
}

/*
获取程序运行路径
*/
func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func connect() *websocket.Conn {
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		fmt.Println(err)
	}

	return ws
}

func heart(ws *websocket.Conn) {

	timer := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-timer.C:
			{
				ws.Write([]byte("{}"))
			}

		}
	}
}
