package box

import (
	"errors"
	"net"
	"syscall"

	"github.com/oswaldoooo/app/internal/linux"
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

type BoxNsConfig struct {
	MountProc string
	Flags     int
}

func (b *BoxConfig) Validate() error {
	if len(b.LinkNet) > 0 {
		linux.Parent_NETNS_ID = b.LinkNet
	}
	if len(b.LinkFs) > 0 {
		linux.Parent_MNTNS_ID = b.LinkFs
	}
	if len(b.LinkPid) > 0 {
		linux.Parent_PIDNS_ID = b.LinkPid
	}
	if len(b.LinkUts) > 0 {
		linux.Parent_UTSNS_ID = b.LinkUts
	}
	return nil
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
