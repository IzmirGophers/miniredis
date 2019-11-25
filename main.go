package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/semihalev/log"
)

var (
	host, port, dbFileName string
	bgSaveInterval         time.Duration
	appLog                 log.Logger
)

const (
	errParamNotEnough       = "Param not enough (required %b)"
	infoDbLoadings          = "DB Loading"
	infoDBFileDoesNotExists = "DB file does not exists, creating (%s)"
	infoDBFileOpening       = "DB file opening (%s)"
	infoTCPListening        = "TCP Listening (%s)"
)

const (
	ver = 1.0
)

type kv struct {
	Key   string `json:"key"`
	Value string `json:"val"`
}

type store struct {
	sync.RWMutex
	l map[string]string
}

func init() {
	flag.StringVar(&host, "host", "127.0.0.1", "Host")
	flag.StringVar(&port, "port", "1234", "Port")
	flag.StringVar(&dbFileName, "file", "store.db", "Store DB filename")
	flag.DurationVar(&bgSaveInterval, "bgSaveInterval", 5, "Background save interval")
}

func main() {
	flag.Parse()

	appLog = log.New("host", host, "port", port, "dbfile", dbFileName, "ver", ver)
	appLog.Info("oOoOo miniredis oOoO")

	memStore := &store{}
	memStore.l = make(map[string]string)

	commands := map[string]func(*store, net.Conn, []string){
		"GET": get,
		"SET": set,
		"DEL": del,
	}

	if !fileExists(dbFileName) {
		appLog.Info(fmt.Sprintf(infoDBFileDoesNotExists, dbFileName))

		_, err := os.Create(dbFileName)

		if err != nil {
			appLog.Error(err.Error())
		}
	}

	appLog.Info(fmt.Sprintf(infoDBFileOpening, dbFileName))
	dbFile, err := os.OpenFile(dbFileName, os.O_RDWR, 0644)

	if err != nil {
		appLog.Error(err.Error())
	}

	go bgSave(memStore, dbFile)
	loadDB(memStore)

	hostURI := fmt.Sprintf("%s:%s", host, port)
	tcpServ, err := net.Listen("tcp", hostURI)

	if err != nil {
		appLog.Error(err.Error())
	}

	defer tcpServ.Close()

	appLog.Info(fmt.Sprintf(infoTCPListening, hostURI))

	conn, err := tcpServ.Accept()

	if err != nil {
		appLog.Error(err.Error())
	}

	for {
		msg, _ := bufio.NewReader(conn).ReadString('\n')
		params := strings.Fields(msg)

		if len(params) < 1 {
			appLog.Error(errParamNotEnough)
		}

		cmd, exists := commands[strings.ToUpper(params[0])]

		if !exists {
			conn.Write([]byte("UNKNONW\n"))
			break
		}

		cmd(memStore, conn, params)
	}
}

//cmds
func get(store *store, conn net.Conn, params []string) {
	if len(params) < 2 {
		appLog.Error(fmt.Sprintf(errParamNotEnough, 1))
		return
	}

	store.RLock()
	conn.Write([]byte(store.l[params[1]] + "\n"))
	store.RUnlock()

}

func del(store *store, conn net.Conn, params []string) {
	if len(params) < 2 {
		appLog.Error(fmt.Sprintf(errParamNotEnough, 1))
		return
	}

	store.RLock()
	_, exists := store.l[params[1]]
	if exists {
		delete(store.l, params[1])
	}
	store.RUnlock()

	conn.Write([]byte("OK\n"))
}

func set(store *store, conn net.Conn, params []string) {
	if len(params) < 3 {
		appLog.Error(fmt.Sprintf(errParamNotEnough, 2))
		return
	}

	store.RLock()
	store.l[params[1]] = params[2]
	store.RUnlock()

	conn.Write([]byte("OK\n"))
}

//bgSave background save function
func bgSave(s *store, f *os.File) {
	for {
		time.Sleep(bgSaveInterval * time.Second)
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
			appLog.Error(err.Error())
		}

		_, err = f.WriteAt(storeRAW, 0)

		if err != nil {
			appLog.Error(err.Error())
		}

	}
}

func loadDB(s *store) {
	appLog.Info(infoDbLoadings)
	data, err := ioutil.ReadFile(dbFileName)

	if err != nil {
		appLog.Error(err.Error())
	}

	var kv = []kv{}

	err = json.Unmarshal(data, &kv)

	if err != nil {
		appLog.Error(err.Error())
	}

	s.RLock()
	for _, v := range kv {
		s.l[v.Key] = v.Value
	}
	s.RUnlock()
}

//util
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
