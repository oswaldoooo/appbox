package boxd

import (
	"errors"
	"log"
	"net"
	"os"
	"strings"
)

type BoxdConfig struct {
	Bind       string
	StreamBind string
	Log        string
}

func (bcnf *BoxdConfig) Logger(_type string) *log.Logger {
	var loggerpath string
	if len(bcnf.Log) == 0 {
		loggerpath = "/var/log/appbox/boxd" + _type + ".log"
	} else {
		index := strings.Index(bcnf.Log, ".log")
		loggerpath = bcnf.Log[:index] + _type + ".log"
	}
	f, err := os.OpenFile(loggerpath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("open logger error", bcnf.Log, err)
	}
	return log.New(f, "", log.Lshortfile|log.Ltime)
}
func (bcnf *BoxdConfig) Validate() error {
	_, err := net.ResolveTCPAddr("tcp4", bcnf.Bind)
	if err != nil {
		return errors.New("api listener bind address error " + err.Error())
	}
	_, err = net.ResolveTCPAddr("tcp4", bcnf.StreamBind)
	if err != nil {
		return errors.New("io stream bind address error " + err.Error())
	}
	return nil
}
