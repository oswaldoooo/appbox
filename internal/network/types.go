package network

import (
	"errors"
	"net"
	"strconv"
)

type NetConfig struct {
	IP       net.IP
	Name     string
	NsPid    string // if it's private. store it's namespace pid
	SubNet   []NetConfig
	Type     NetType
	BrdAttr  BrdAttr
	VethAttr VethAttr
}
type NetType uint

func (t NetType) IsValid() bool {
	return t != Invalid && t <= Tap
}

const (
	Invalid NetType = iota
	Brd
	Veth
	Tun
	Tap
)

var net_type_map = []string{
	"Invalid",
	"bridge",
	"veth",
	"tun",
	"tap",
}

type BrdAttr struct {
	Name string
	IP   net.IP
}
type VethPair struct {
	Name  string
	IP    net.IP
	NsPid string
}
type VethAttr struct {
	PairA VethPair
	PairB VethPair
}

func (nc NetConfig) IsBridge() bool {
	return nc.Type&Brd > 0
}
func (nc NetConfig) IsVeth() bool {
	return nc.Type&Veth > 0
}
func (nc *NetConfig) SetName(name string) {
	nc.Name = name
	if nc.IsVeth() {
		nc.VethAttr.PairA.Name = name
	}
}
func (t *NetType) UnmarshalText(text []byte) error {
	for i, v := range net_type_map {
		if v == string(text) {
			*t = NetType(i)
			return nil
		}
	}
	return errors.New("invalid type " + string(text))
}
func (t NetType) MarshalText() (text []byte, err error) {
	text = []byte(t.String())
	return
}

func (t *NetType) String() string {
	if int(*t) >= len(net_type_map) {
		return "invalid net type " + strconv.Itoa(int(*t))
	}
	return net_type_map[*t]
}

func (t *NetType) Set(s string) error {
	return t.UnmarshalText([]byte(s))
}

func (t *NetType) Type() string {
	return "net-type"
}
