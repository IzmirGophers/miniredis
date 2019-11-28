package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"net"
	"os"
	"reflect"
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
	commands               map[string]func(*hashmap.HashMap, []string) []string
	store                  *hashmap.HashMap
)

const (
	errEmptyMsg         = "You need to use one of these: %s"
	errParamNotEnough   = "Param not enough (required %d)"
	infoDbLoadings      = "DB Loading"
	infoDBFileOpening   = "DB file opening (%s)"
	infoTCPListening    = "TCP Listening (%s)"
	infoClientConnected = "Client connected (%s)"

	//default response
	responseNull = "NULL\n"
	responseOK   = "OK\n"
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

	commands = map[string]func(*hashmap.HashMap, []string) []string{
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
		msg, err := bufio.NewReader(c).ReadString('\n')

		if err != nil {
			continue
		}

		if len(msg) < 3 {
			keys := reflect.ValueOf(commands).MapKeys()
			errMsg := fmt.Sprintf(errEmptyMsg, keys)
			c.Write([]byte(errMsg + "\n"))
			appLog.Info(errMsg)
			continue
		}

		params := strings.Fields(msg)

		cmd, exists := commands[strings.ToUpper(params[0])]

		if !exists {
			keys := reflect.ValueOf(commands).MapKeys()
			errMsg := fmt.Sprintf(errEmptyMsg, keys)
			c.Write([]byte(errMsg + "\n"))
			appLog.Info(errMsg)
			continue
		}

		rsp := cmd(store, params)

		for _, r := range rsp {
			c.Write([]byte(r + "\n"))
		}

	}
}

//cmds
func get(s *hashmap.HashMap, p []string) (rsp []string) {
	if len(p) < 2 {
		appLog.Error(fmt.Sprintf(errParamNotEnough, 1))
		return rsp
	}

	val, ok := s.Get(p[1])

	if !ok {
		rsp = append(rsp, responseNull)
		return rsp
	}

	rsp = append(rsp, val.(string))

	return rsp
}

func mget(s *hashmap.HashMap, p []string) (rsp []string) {
	if len(p) < 2 {
		appLog.Error(fmt.Sprintf(errParamNotEnough, 1))
		return rsp
	}

	for i := 1; i < len(p); i++ {

		val, ok := s.Get(p[i])
		if !ok {
			rsp = append(rsp, responseNull)
			return rsp
		}

		rsp = append(rsp, val.(string))

	}

	return rsp
}

func del(s *hashmap.HashMap, p []string) (rsp []string) {
	if len(p) < 2 {
		appLog.Error(fmt.Sprintf(errParamNotEnough, 1))
		return rsp
	}

	_, exists := s.Get(p[1])

	if exists {
		s.Del(p[1])
	}

	rsp = append(rsp, responseOK)
	return rsp
}

func set(s *hashmap.HashMap, p []string) (rsp []string) {
	if len(p) < 3 {
		appLog.Error(fmt.Sprintf(errParamNotEnough, 2))
		return rsp
	}

	s.Set(p[1], p[2])

	rsp = append(rsp, responseOK)
	return rsp
}

func mset(s *hashmap.HashMap, p []string) (rsp []string) {
	if len(p) < 3 || (len(p)-1)%2 == 1 {
		appLog.Error(fmt.Sprintf(errParamNotEnough, 2))
		return rsp
	}

	for i := 1; i < len(p); i += 2 {
		s.Set(p[i], p[i+1])
	}

	rsp = append(rsp, responseOK)
	return rsp
}

func dbSize(s *hashmap.HashMap, p []string) (rsp []string) {
	length := strconv.Itoa(s.Len())
	rsp = append(rsp, length)
	return rsp
}

func keys(s *hashmap.HashMap, p []string) (rsp []string) {

	for item := range s.Iter() {
		rsp = append(rsp, item.Key.(string))
	}
	return rsp
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
