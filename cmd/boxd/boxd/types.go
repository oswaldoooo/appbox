package boxd

import (
	"errors"
	"net"
)

type BoxdConfig struct {
	Bind       string
	StreamBind string
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
