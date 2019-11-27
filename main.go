package main

import (
	"bufio"
	"encoding/gob"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/semihalev/log"
)

var (
	host, port, dbFileName string
	bgSaveInterval         time.Duration
	appLog                 log.Logger
	commands               map[string]func(*store, net.Conn, []string)
	memStore               *store
)

const (
	errParamNotEnough       = "Param not enough (required %d)"
	infoDbLoadings          = "DB Loading"
	infoDBFileDoesNotExists = "DB file does not exists, creating (%s)"
	infoDBFileOpening       = "DB file opening (%s)"
	infoTCPListening        = "TCP Listening (%s)"
	infoClientConnected     = "Client connected (%s)"

	//default response
	responseNull    = "NULL\n"
	responseOK      = "OK\n"
	responseUnknown = "UNKNOWN\n"
)

const (
	ver = 1.0
)

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

	memStore = &store{}
	memStore.l = make(map[string]string)

	commands = map[string]func(*store, net.Conn, []string){
		"GET":    get,
		"SET":    set,
		"DEL":    del,
		"DBSIZE": dbSize,
	}

	if !fileExists(dbFileName) {
		appLog.Info(fmt.Sprintf(infoDBFileDoesNotExists, dbFileName))

		_, err := os.Create(dbFileName)

		if err != nil {
			appLog.Error(err.Error())
			os.Exit(1)
		}
	}

	appLog.Info(fmt.Sprintf(infoDBFileOpening, dbFileName))
	dbFile, err := os.OpenFile(dbFileName, os.O_RDWR, 0644)

	if err != nil {
		appLog.Error(err.Error())
		os.Exit(1)
	}

	loadDB(memStore, dbFile)
	go bgSave(memStore, gob.NewEncoder(dbFile))

	hostURI := fmt.Sprintf("%s:%s", host, port)
	tcpServ, err := net.Listen("tcp", hostURI)

	if err != nil {
		appLog.Error(err.Error())
		os.Exit(1)
	}

	defer tcpServ.Close()

	appLog.Info(fmt.Sprintf(infoTCPListening, hostURI))

	for {
		conn, err := tcpServ.Accept()

		if err != nil {
			appLog.Error(err.Error())
			os.Exit(1)
		}

		go listen(conn)
	}

}

func listen(c net.Conn) {
	appLog.Info(fmt.Sprintf(infoClientConnected, c.RemoteAddr()))
	for {
		msg, _ := bufio.NewReader(c).ReadString('\n')
		params := strings.Fields(msg)

		if len(params) < 1 {
			appLog.Error(errParamNotEnough)
		}

		cmd, exists := commands[strings.ToUpper(params[0])]

		if !exists {
			c.Write([]byte(responseUnknown))
			continue
		}

		cmd(memStore, c, params)
	}
}

//cmds
func get(s *store, c net.Conn, p []string) {
	if len(p) < 2 {
		appLog.Error(fmt.Sprintf(errParamNotEnough, 1))
		return
	}

	s.RLock()
	val, ok := s.l[p[1]]

	if !ok {
		c.Write([]byte(responseNull))
		return
	}
	c.Write([]byte(val + "\n"))
	s.RUnlock()

}

func del(s *store, c net.Conn, p []string) {
	if len(p) < 2 {
		appLog.Error(fmt.Sprintf(errParamNotEnough, 1))
		return
	}

	s.RLock()
	_, exists := s.l[p[1]]
	if exists {
		delete(s.l, p[1])
	}
	s.RUnlock()

	c.Write([]byte(responseOK))
}

func set(s *store, c net.Conn, p []string) {
	if len(p) < 3 {
		appLog.Error(fmt.Sprintf(errParamNotEnough, 2))
		return
	}

	s.RLock()
	s.l[p[1]] = p[2]
	s.RUnlock()

	c.Write([]byte(responseOK))
}

func dbSize(s *store, c net.Conn, p []string) {
	length := strconv.Itoa(len(s.l))
	c.Write([]byte(length + "\n"))
}

//bgSave background save function
func bgSave(s *store, enc *gob.Encoder) {
	for {
		time.Sleep(bgSaveInterval * time.Second)

		if err := enc.Encode(s.l); err != nil {
			panic(err)
		}

	}
}

func loadDB(s *store, dbFile *os.File) {
	appLog.Info(infoDbLoadings)

	decoder := gob.NewDecoder(dbFile)
	decoder.Decode(&s.l)
}

//util
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
