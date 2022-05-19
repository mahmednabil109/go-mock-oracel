package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"sync"
	"time"

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
	Hamda       float32
}

type Variable struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

// type Config struct {
// 	Name      string     `json:"name"`
// 	Variables []Variable `json:"variables"`
// }

// func read_config() (*Config, error) {
// 	file, err := ioutil.ReadFile(*CONFIG_PATH)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var config Config
// 	err = json.Unmarshal(file, &config)

// 	return &config, err
// }
type Update struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	Timestamp string `json:"timestamp"`
}

type Queue struct {
	Modle []Variable
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
		field := mType.Field(i)
		q.Modle = append(q.Modle, Variable{Name: field.Name, Type: field.Type.String()})
		q.Topic[field.Name] = []*websocket.Conn{}
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

func (q *Queue) Update(v Variable) {
	conns, ok := q.Topic[v.Name]
	if !ok {
		return
	}

	for _, conn := range conns {
		go func(conn *websocket.Conn) {
			res := Update{
				Key:       v.Name,
				Value:     v.Value,
				Timestamp: time.Now().String(),
			}

			err := conn.WriteJSON(&res)
			if err != nil {
				log.Print("update faild :(")
			}
		}(conn)
	}
}

// MOCK functions

// func (q *Queue) UpdateTemperature(value float64) {
// 	conns := q.Topic["Temperature"]
// 	for _, conn := range conns {
// 		go func(conn *websocket.Conn) {
// 			res := Update{
// 				Key:       "Temperature",
// 				Value:     fmt.Sprint(value),
// 				Timestamp: time.Now().String(),
// 			}

// 			err := conn.WriteJSON(&res)
// 			if err != nil {
// 				log.Print("update faild :(")
// 			}
// 		}(conn)
// 	}
// }

// func (q *Queue) UpdateRicePrice(value float64) {
// 	conns := q.Topic["RicePrice"]
// 	for _, conn := range conns {
// 		go func(conn *websocket.Conn) {
// 			res := Update{
// 				Key:       "RicePrice",
// 				Value:     fmt.Sprint(value),
// 				Timestamp: time.Now().String(),
// 			}

// 			err := conn.WriteJSON(&res)
// 			if err != nil {
// 				log.Print("update faild :(")
// 			}
// 		}(conn)
// 	}
// }

func init_server(q *Queue) {
	http.HandleFunc("/ui", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./ui.html")
	})

	http.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		var v Variable
		err := json.NewDecoder(r.Body).Decode(&v)
		if err != nil {
			return
		}

		log.Printf("%+v", v)
		q.Update(v)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(struct{ Msg string }{"done"})
	})

	http.HandleFunc("/models", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err := json.NewEncoder(w).Encode(q.Modle)
		if err != nil {
			panic(err)
		}
	})

	http.HandleFunc("/sub", q.HandleSub)

	log.Fatal(http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", *PORT), nil))
}

func main() {
	flag.Parse()

	var queue Queue
	queue.Init()
}
