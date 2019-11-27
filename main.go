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
	errParamNotEnough   = "Param not enough (required %d)"
	infoDbLoadings      = "DB Loading"
	infoDBFileOpening   = "DB file opening (%s)"
	infoTCPListening    = "TCP Listening (%s)"
	infoClientConnected = "Client connected (%s)"

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
	flag.DurationVar(&bgSaveInterval, "bgSaveInterval", 1, "Background save interval")
}

func main() {
	flag.Parse()

	appLog = log.New("host", host, "port", port, "dbfile", dbFileName, "ver", ver)
	appLog.Info("oOoOo miniredis oOoO")

	memStore = &store{}
	memStore.l = make(map[string]string)

	commands = map[string]func(*store, net.Conn, []string){
		"GET":    get,
		"MGET":   mget,
		"SET":    set,
		"MSET":   mset,
		"DEL":    del,
		"DBSIZE": dbSize,
		"KEYS":   keys,
	}

	appLog.Info(fmt.Sprintf(infoDBFileOpening, dbFileName))
	dbFile, err := os.OpenFile(dbFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)

	if err != nil {
		appLog.Error(err.Error())
		os.Exit(1)
	}

	err = loadDB(memStore, dbFile)

	if err != nil {
		// handle
	}

	go bgSave(memStore, dbFile)

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

func mget(s *store, c net.Conn, p []string) {
	if len(p) < 2 {
		appLog.Error(fmt.Sprintf(errParamNotEnough, 1))
		return
	}

	s.RLock()
	for i := 1; i < len(p); i++ {

		val, ok := s.l[p[i]]
		if !ok {
			c.Write([]byte(responseNull))
		}
		c.Write([]byte(val + "\n"))

	}

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

func mset(s *store, c net.Conn, p []string) {
	if len(p) < 3 || (len(p)-1)%2 == 1 {
		appLog.Error(fmt.Sprintf(errParamNotEnough, 2))
		return
	}

	s.RLock()
	for i := 1; i < len(p); i += 2 {
		s.l[p[i]] = p[i+1]

	}
	s.RUnlock()

	c.Write([]byte(responseOK))
}

func dbSize(s *store, c net.Conn, p []string) {
	length := strconv.Itoa(len(s.l))
	c.Write([]byte(length + "\n"))

}

func keys(s *store, c net.Conn, p []string) {
	s.RLock()
	for key, _ := range s.l {
		c.Write([]byte(key + "\n"))
	}
	s.RUnlock()
}

//bgSave background save function
func bgSave(s *store, f *os.File) {
	for {
		time.Sleep(bgSaveInterval * time.Second)
		//bbgsave olayını msgpack al
		f.Truncate(0)

		enc := gob.NewEncoder(f)

		if err := enc.Encode(s.l); err != nil {
			panic(err)
		}

	}
}

func loadDB(s *store, f *os.File) error {
	appLog.Info(infoDbLoadings)

	decoder := gob.NewDecoder(f)
	err := decoder.Decode(&s.l)

	if err != nil {
		return err
	}

	return nil
}
