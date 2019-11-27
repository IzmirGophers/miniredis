package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cornelk/hashmap"
	"github.com/semihalev/log"
)

var (
	host, port, dbFileName string
	bgSaveInterval         time.Duration
	appLog                 log.Logger
	commands               map[string]func(*hashmap.HashMap, net.Conn, []string)
	store                  *hashmap.HashMap
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

	store = &hashmap.HashMap{}

	commands = map[string]func(*hashmap.HashMap, net.Conn, []string){
		"GET":    get,
		"MGET":   mget,
		"SET":    set,
		"MSET":   mset,
		"DEL":    del,
		"DBSIZE": dbSize,
		"KEYS":   keys,
	}

	appLog.Info(fmt.Sprintf(infoDBFileOpening, dbFileName))
	dbFile, err := os.OpenFile(dbFileName, os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {
		appLog.Error(err.Error())
		os.Exit(1)
	}

	loadDB(store, dbFile)

	go bgSave(store, dbFile)

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

		cmd(store, c, params)
	}
}

//cmds
func get(s *hashmap.HashMap, c net.Conn, p []string) {
	if len(p) < 2 {
		appLog.Error(fmt.Sprintf(errParamNotEnough, 1))
		return
	}

	val, ok := s.Get(p[1])

	if !ok {
		c.Write([]byte(responseNull))
		return
	}
	c.Write([]byte(val.(string) + "\n"))
}

func mget(s *hashmap.HashMap, c net.Conn, p []string) {
	if len(p) < 2 {
		appLog.Error(fmt.Sprintf(errParamNotEnough, 1))
		return
	}

	for i := 1; i < len(p); i++ {

		val, ok := s.Get(p[i])
		if !ok {
			c.Write([]byte(responseNull))
		}
		c.Write([]byte(val.(string) + "\n"))
	}

}

func del(s *hashmap.HashMap, c net.Conn, p []string) {
	if len(p) < 2 {
		appLog.Error(fmt.Sprintf(errParamNotEnough, 1))
		return
	}

	_, exists := s.Get(p[1])

	if exists {
		s.Del(p[1])
	}

	c.Write([]byte(responseOK))
}

func set(s *hashmap.HashMap, c net.Conn, p []string) {
	if len(p) < 3 {
		appLog.Error(fmt.Sprintf(errParamNotEnough, 2))
		return
	}
	s.Set(p[1], p[2])
	c.Write([]byte(responseOK))
}

func mset(s *hashmap.HashMap, c net.Conn, p []string) {
	if len(p) < 3 || (len(p)-1)%2 == 1 {
		appLog.Error(fmt.Sprintf(errParamNotEnough, 2))
		return
	}

	for i := 1; i < len(p); i += 2 {
		s.Set(p[i], p[i+1])
	}

	c.Write([]byte(responseOK))
}

func dbSize(s *hashmap.HashMap, c net.Conn, p []string) {
	length := strconv.Itoa(s.Len())
	c.Write([]byte(length + "\n"))

}

func keys(s *hashmap.HashMap, c net.Conn, p []string) {

	for item := range s.Iter() {
		c.Write([]byte(item.Value.(string) + "\n"))
	}

}

//bgSave background save function
func bgSave(s *hashmap.HashMap, f *os.File) {
	for {
		time.Sleep(bgSaveInterval * time.Second)

		var (
			buf bytes.Buffer
			tmp = make(map[string]string)
		)

		for item := range s.Iter() {
			tmp[item.Key.(string)] = item.Value.(string)
		}

		enc := gob.NewEncoder(&buf)

		if err := enc.Encode(tmp); err != nil {
			appLog.Error(err.Error())
		}

		f.WriteAt(buf.Bytes(), 0)
	}
}

func loadDB(s *hashmap.HashMap, f *os.File) {
	appLog.Info(infoDbLoadings)

	var tmp = make(map[string]string)

	decoder := gob.NewDecoder(f)
	decoder.Decode(&tmp)

	for key, value := range tmp {
		s.Set(key, value)
	}

}
