package box

import (
	"bytes"
	"errors"
	"net"
	"strconv"
	"syscall"

	"github.com/oswaldoooo/app/internal/network"
)

type BoxConfig struct {
	StaticPath     []string
	Path           []string
	Run            []string
	StandloneNet   bool
	StandloneFs    bool
	StandloneUsers bool
	StandloneHost  bool
	//config mode
	LinkMode bool
	LinkNet  string
	LinkFs   string
	LinkPid  string
	LinkUts  string
}

func (b BoxConfig) Valid() error {
	if b.StandloneNet && len(b.LinkNet) > 0 {
		return errors.New("namespace mode conflicted on net")
	}
	if b.StandloneFs && len(b.LinkFs) > 0 {
		return errors.New("namespace mode conflicted on filesystem")
	}
	if b.StandloneHost && len(b.LinkUts) > 0 {
		return errors.New("namespace mode conflicted on uts")
	}
	return nil
}

type BoxNsConfig struct {
	MountProc string
	Flags     int
}

func (b BoxConfig) NsConfig() (nscnf BoxNsConfig) {
	flags := syscall.CLONE_NEWPID
	if b.StandloneFs {
		flags |= syscall.CLONE_NEWNS
	}
	if b.StandloneNet {
		flags |= syscall.CLONE_NEWNET
	}
	if b.StandloneUsers {
		flags |= syscall.CLONE_NEWUSER
	}
	if b.StandloneHost {
		flags |= syscall.CLONE_NEWUTS
	}
	nscnf.Flags = flags
	nscnf.MountProc = "/app"
	return
}

type BoxNetConfig struct {
	Action           string //must require
	IP               net.IP
	Name             string
	Pid              string
	Type             network.NetType
	network.BrdAttr  `json:",omitempty"`
	network.VethAttr `json:",omitempty"`
}

func (b *BoxNetConfig) Valid() error {
	if b.Action == "add" && b.IP == nil && len(b.IP) != 4 && len(b.IP) != 8 {
		return errors.New("invalid ip address")
	} else if b.Action != "delete" && !b.Type.IsValid() {
		return errors.New("invalid net type")
	} else if len(b.Action) == 0 {
		return errors.New("not set action")
	}
	b.VethAttr.PairB.IP = b.IP
	return nil
}
func (b *BoxNetConfig) IsVaild() bool {
	return b.Valid() == nil
}

type IP struct {
	IP      net.IP
	KeepBit uint8
}

func (i *IP) String() string {
	return "127.0.0.1/24"
}

func (i *IP) Set(s string) error {
	return i.UnmarshalText([]byte(s))
}

func (i *IP) Type() string {
	return "ip"
}
func (i *IP) UnmarshalText(text []byte) error {
	index := bytes.IndexByte(text, '/')
	if index < 0 {
		return errors.New("format error")
	}
	err := i.IP.UnmarshalText(text[:index])
	if err != nil {
		return err
	}
	bit, err := strconv.ParseUint(string(text[index+1:]), 10, 8)
	if err != nil {
		return err
	}
	if bit > 32 {
		return errors.New("netmask is 0~32")
	}
	i.KeepBit = uint8(bit)
	return nil
}
