package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"sync"
)

type Message struct {
	Message string ``
}

var (
	upgrader = websocket.Upgrader{
		// 允许跨域
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan Message)

func execFunc(cmd *exec.Cmd, conn *Connection) {

	//显示运行的命令

	stdout, err := cmd.StdoutPipe()

	if err != nil {
		fmt.Println(err)
	}
	cmd.Start()

	reader := bufio.NewReader(stdout)

	//实时循环读取输出流中的一行内容
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			fmt.Println(err2)
			break
		}
		conn.WriteMessage([]byte(line))

	}
	cmd.Wait()
}
func execHandle(cmdString string, conn *Connection) {
	cmdString = strings.TrimSpace(cmdString)
	var rawStrings = strings.SplitN(cmdString, " ", 2)
	var cmdStr = rawStrings[0]
	var flgList = ""
	if len(rawStrings) > 1 {
		flgList = rawStrings[1]
		var flags = strings.Split(flgList, " ")

		cmd := exec.Command(cmdStr, flags...)
		execFunc(cmd, conn)
	} else {
		cmd := exec.Command(cmdStr)
		execFunc(cmd, conn)
	}

}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	//	w.Write([]byte("hello"))
	var (
		wsConn *websocket.Conn
		err    error
		conn   *Connection
		data   []byte
	)
	// 完成ws协议的握手操作
	// Upgrade:websocket
	if wsConn, err = upgrader.Upgrade(w, r, nil); err != nil {
		return
	}

	if conn, err = initConnection(wsConn); err != nil {
		goto ERR
	}

	for {
		if data, err = conn.ReadMessage(); err != nil {
			goto ERR
		}
		stringData := string(data)
		execHandle(stringData, conn)
		/*		err = conn.WriteMessage(data)
				if err != nil {
					goto ERR
				}*/
	}

ERR:
	conn.Close()

}

type Connection struct {
	wsConnect *websocket.Conn
	inChan    chan []byte
	outChan   chan []byte
	closeChan chan byte
	mutex     sync.Mutex // 对closeChan关闭上锁
	isClosed  bool       // 防止closeChan被关闭多次
}

func initConnection(wsConn *websocket.Conn) (conn *Connection, err error) {
	conn = &Connection{
		wsConnect: wsConn,
		inChan:    make(chan []byte, 1000),
		outChan:   make(chan []byte, 1000),
		closeChan: make(chan byte, 1),
	}
	// 启动读协程
	go conn.readLoop()
	// 启动写协程
	go conn.writeLoop()
	return
}

func (conn *Connection) ReadMessage() (data []byte, err error) {

	select {
	case data = <-conn.inChan:
		/*fmt.Printf(string(data[:]))*/
	case <-conn.closeChan:
		err = errors.New("connection is closeed")
	}
	return
}

func (conn *Connection) WriteMessage(data []byte) (err error) {

	select {
	case conn.outChan <- data:
	case <-conn.closeChan:
		err = errors.New("connection is closeed")
	}
	return
}

func (conn *Connection) Close() {
	// 线程安全，可多次调用
	conn.wsConnect.Close()
	// 利用标记，让closeChan只关闭一次
	conn.mutex.Lock()
	if !conn.isClosed {
		close(conn.closeChan)
		conn.isClosed = true
	}
	conn.mutex.Unlock()
}

// 内部实现
func (conn *Connection) readLoop() {
	var (
		data []byte
		err  error
	)
	for {
		if _, data, err = conn.wsConnect.ReadMessage(); err != nil {
			goto ERR
		}
		//阻塞在这里，等待inChan有空闲位置
		select {
		case conn.inChan <- data:
		case <-conn.closeChan: // closeChan 感知 conn断开
			goto ERR
		}

	}

ERR:
	conn.Close()
}

func (conn *Connection) writeLoop() {
	var (
		data []byte
		err  error
	)

	for {
		select {
		case data = <-conn.outChan:
		case <-conn.closeChan:
			goto ERR
		}
		if err = conn.wsConnect.WriteMessage(websocket.TextMessage, data); err != nil {
			goto ERR
		}
	}

ERR:
	conn.Close()

}
