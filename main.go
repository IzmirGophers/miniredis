package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	host, port, dbFileName string
	store                  map[string]string
)

type kv struct {
	Key   string `json:"key"`
	Value string `json:"val"`
}

type Store struct {
	sync.RWMutex
	l map[string]string
}

func init() {
	flag.StringVar(&host, "host", "127.0.0.1", "host")
	flag.StringVar(&port, "port", "1234", "port")
	flag.StringVar(&dbFileName, "file", "store.db", "file")
}

func main() {
	flag.Parse()

	var store = &Store{}
	store.l = make(map[string]string)

	if !fileExists(dbFileName) {
		_, err := os.Create(dbFileName)

		if err != nil {
			log.Fatal("Error creating DBfile:", err.Error())
		}
	}

	dbFile, err := os.OpenFile(dbFileName, os.O_RDWR, 0644)

	if err != nil {
		log.Fatal("dbfile can not open")
	}

	go bgSave(store, dbFile)
	loadDB(store)

	hostURI := fmt.Sprintf("%s:%s", host, port)

	l, err := net.Listen("tcp", hostURI)

	if err != nil {
		log.Fatal("Error listening:", err.Error())
	}

	defer l.Close()

	fmt.Println("Listening on ", hostURI)

	// Listen for an incoming connection.
	conn, err := l.Accept()

	if err != nil {
		log.Fatal("Error accepting: ", err.Error())
	}

	for {
		msg, _ := bufio.NewReader(conn).ReadString('\n')

		params := strings.Fields(msg)

		if len(params) < 1 {
			fmt.Println("param not enough")
		}

		switch params[0] {
		case "SET", "set":
			if len(params) < 3 {
				fmt.Println("param not enough")
				break
			}

			store.RLock()
			store.l[params[1]] = params[2]
			store.RUnlock()

			conn.Write([]byte("OK\n"))
			break
		case "GET", "get":
			if len(params) < 2 {
				fmt.Println("param not enough")
				break
			}
			store.RLock()
			conn.Write([]byte(store.l[params[1]] + "\n"))
			store.RUnlock()

			break
		case "DEL", "del":

			if len(params) < 2 {
				fmt.Println("param not enough")
				break
			}

			store.RLock()
			_, exists := store.l[params[1]]

			if exists {
				delete(store.l, params[1])
			}
			store.RUnlock()

			conn.Write([]byte("OK\n"))
			break
		default:
			conn.Write([]byte("UNKNOWN\n"))
			break
		}
	}

}

//bgSave background save function
func bgSave(s *Store, f *os.File) {
	for {
		time.Sleep(1 * time.Second)
		var storedb = []kv{}

		s.RLock()
		for k, v := range s.l {
			storedb = append(storedb, kv{
				Key:   k,
				Value: v,
			})
		}
		s.RUnlock()

		storeRAW, err := json.Marshal(storedb)

		if err != nil {
			log.Fatal("db bgsave error (serializing) - ", err.Error())
		}

		_, err = f.WriteAt(storeRAW, 0)

		if err != nil {
			log.Fatal("db bgsave error (write) - ", err.Error())
		}

	}
}

func loadDB(s *Store) {
	data, err := ioutil.ReadFile(dbFileName)

	if err != nil {
		log.Fatal("loaddb error - ", err.Error())
	}

	var kv = []kv{}

	err = json.Unmarshal(data, &kv)

	if err != nil {
		log.Fatal("loadd error - ", err.Error())
	}

	s.RLock()
	for _, v := range kv {
		s.l[v.Key] = v.Value
	}
	s.RUnlock()
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
