package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	PORT        = flag.Int("port", 8383, "server port")
	CONFIG_PATH = flag.String("config", "./config.json", "config file path")
	upgrader    = websocket.Upgrader{}
)

// mocking model
// TODO replace this struct using reflect technique
type Model struct {
	Temperature float32
	RicePrice   float32
}

// type Variable struct {
// 	Name string `json:"name"`
// 	Type string `json:"type"`
// }

// type Config struct {
// 	Name      string     `json:"name"`
// 	Variables []Variable `json:"variables"`
// }

type Queue struct {
	Topic map[string][]*websocket.Conn
	Mux   sync.Mutex
}

func (q *Queue) Init() {
	// for _, v := range cfg.Variables {
	// 	q.Topic[v.Name] = []*websocket.Conn{}
	// }

	q.Topic = make(map[string][]*websocket.Conn)
	mType := reflect.TypeOf(Model{})
	for i := 0; i < mType.NumField(); i++ {
		log.Print(mType.Field(i).Name)
		q.Topic[mType.Field(i).Name] = []*websocket.Conn{}
	}

	init_server(q)
}

func (q *Queue) HandleSub(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
	log.Print("As")

	// TODO read json req
	_, variable, err := conn.ReadMessage()
	log.Print(string(variable))
	if err != nil {
		panic(err)
	}

	_, ok := q.Topic[string(variable)]
	log.Print(ok)
	if !ok {
		return
	}

	q.Mux.Lock()
	defer q.Mux.Unlock()
	q.Topic[string(variable)] = append(q.Topic[string(variable)], conn)

	// assume that the other end will close the connections :)
}

func (q *Queue) UpdateTemperature(value float64) {
	// conns := q.Topic["Temperature"]
	// for _, conn := range conns {
	// TODO send json response
	// conn
	// }
}

// func read_config() (*Config, error) {
// 	file, err := ioutil.ReadFile(*CONFIG_PATH)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var config Config
// 	err = json.Unmarshal(file, &config)

// 	return &config, err
// }

func init_server(q *Queue) {
	http.HandleFunc("/ui", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./ui.html")
	})

	http.HandleFunc("/sub", q.HandleSub)

	log.Fatal(http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", *PORT), nil))
}

func main() {
	flag.Parse()

	// config, err := read_config()
	// if err != nil {
	// 	panic(err)
	// }

	var queue Queue
	queue.Init()
}
