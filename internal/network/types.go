package network

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net"
	"strconv"
)

type NetConfig struct {
	IP       net.IP
	Name     string
	KeepBit  uint8
	NsPid    string // if it's private. store it's namespace pid
	SubNet   []NetConfig
	Type     NetType
	BrdAttr  BrdAttr
	VethAttr VethAttr
}

func (nc *NetConfig) Validate() error {
	if !nc.Type.IsValid() {
		return errors.New("net config is invalid")
	}
	if nc.IsBridge() {
		nc.BrdAttr.IP = nc.IP
		nc.BrdAttr.Name = nc.Name
		nc.BrdAttr.KeepBit = nc.KeepBit
	} else if nc.IsVeth() {
		nc.VethAttr.PairA.IP = nc.IP
		nc.VethAttr.PairA.Name = nc.Name
		nc.VethAttr.PairA.KeepBit = nc.KeepBit
		if len(nc.VethAttr.PairB.Name) == 0 {
			var rd [2]byte
			rand.Read(rd[:])
			nc.VethAttr.PairB.Name = "veth" + hex.EncodeToString(rd[:])
		}
	}
	return nil
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
const default_keep_bit = 24

var net_type_map = []string{
	"Invalid",
	"bridge",
	"veth",
	"tun",
	"tap",
}

type IpInterface interface {
	IPString() string
}
type BrdAttr struct {
	Name    string
	IP      net.IP
	KeepBit uint8
}

func (v *VethPair) IPString() string {
	if v.KeepBit == 0 {
		v.KeepBit = default_keep_bit
	}
	return v.IP.String() + "/" + strconv.Itoa(int(v.KeepBit))
}
func (b *BrdAttr) IPString() string {
	if b.KeepBit == 0 {
		b.KeepBit = default_keep_bit
	}
	return b.IP.String() + "/" + strconv.Itoa(int(b.KeepBit))
}

type VethPair struct {
	Name    string
	IP      net.IP
	KeepBit uint8 // ip/keepbit=127.0.0.1/8
	NsPid   string
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

func (nc *NetConfig) SetVeth(name string, addr net.IP, bit uint8, pid string, peer_pid string) {
	nc.IP = addr
	nc.KeepBit = bit
	nc.Name = name
	nc.NsPid = pid
	nc.VethAttr.PairA.Name = name
	nc.VethAttr.PairA.IP = addr
	nc.VethAttr.PairA.KeepBit = bit
	nc.VethAttr.PairA.NsPid = pid
	nc.Type = Veth
	var rd [2]byte
	rand.Read(rd[:])
	nc.VethAttr.PairB.Name = "veth" + hex.EncodeToString(rd[:])
	nc.VethAttr.PairB.NsPid = peer_pid
}

func (nc *NetConfig) SetBrd(name string, addr net.IP, bit uint8, pid string) {
	nc.IP = addr
	nc.KeepBit = bit
	nc.Name = name
	nc.NsPid = pid
	nc.BrdAttr.IP = addr
	nc.BrdAttr.KeepBit = bit
	nc.BrdAttr.Name = name
}
