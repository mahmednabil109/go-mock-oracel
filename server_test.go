package main

import (
	"log"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func Test_sub(t *testing.T) {

	var queue Queue
	go queue.Init()

	time.Sleep(1000 * time.Millisecond)

	conn := getConn()
	conn.WriteMessage(websocket.TextMessage, []byte("Temperature"))
	time.Sleep(1000 * time.Millisecond)

	log.Printf("%+v", queue.Topic)
}
func Test_generic(t *testing.T) {
	tf := func(v interface{}) {
		switch v.(type) {
		case int:
			log.Print("int")
		case string:
			log.Print("Stirng")
		case float64:
			log.Print("float64")
		default:
			log.Println("None")
		}
	}

	tf(1)
	tf(1.2)
	tf("asd")

}

func getConn() *websocket.Conn {
	u := url.URL{Scheme: "ws", Host: "127.0.0.1:8383", Path: "/sub"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		panic(err)
	}
	return c
}
