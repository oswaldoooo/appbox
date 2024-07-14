package network

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/oswaldoooo/app/internal/linux"
)

const (
	nc_path = "/etc/appbox/appbox-net.json"
)

var (
	TargetNs string
)

func getNcPath() string {
	return "/etc/appbox/appbox-net" + TargetNs + ".json"
}
func Lock() error {
	f, err := os.OpenFile(getNcPath()+".lock", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	f.WriteString(strconv.Itoa(os.Getpid()))
	f.Close()
	return nil
}
func Unlock() error {
	return os.Remove(getNcPath() + ".lock")
}
func GetNetConfig() (nc NetConfig, err error) {
	var f *os.File
	f, err = os.OpenFile(getNcPath(), os.O_RDONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&nc)
	return
}

func NewSubnet(cnf *NetConfig) (nc NetConfig, err error) {
	var mainnc NetConfig = *cnf
	if !mainnc.IsBridge() {
		err = errors.New("net is not bridge network")
		return
	}
	var _ip net.IP
	if len(mainnc.SubNet) == 0 {
		_ip = mainnc.IP
	} else {
		_ip = mainnc.SubNet[len(mainnc.SubNet)-1].IP
	}
	nc.IP = make(net.IP, len(_ip))
	copy(nc.IP, _ip)
	nc.IP[len(nc.IP)-1]++
	nc.Type = Veth
	var rdbuff [2]byte
	rand.Read(rdbuff[:])
	nc.Name = "eth0"
	nc.VethAttr = VethAttr{
		PairA: VethPair{
			Name: "eth0",
			IP:   nc.IP,
		},
		PairB: VethPair{
			Name: "veth-" + hex.EncodeToString(rdbuff[:]),
		},
	}
	return
}
func Dump(netcnf *NetConfig) error {
	f, err := os.OpenFile(getNcPath(), os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "    ")
	return encoder.Encode(netcnf)
}
func NewInterface(nc *NetConfig, netns string) error {
	var (
		err  error
		name string
		ip   string
		cmd  []string
	)
	if nc.IsBridge() {
		name = nc.Name
		ip = nc.IP.String()
		cmd = append(cmd, "ip", "link", "add", "dev", name, "type", "bridge")
		err = NetRaw(netns, cmd...)
	} else if nc.IsVeth() {
		name = nc.VethAttr.PairA.Name
		ip = nc.VethAttr.PairA.IP.String()
		cmd = append(cmd, "ip", "link", "add", "dev", name)
		if len(nc.VethAttr.PairA.NsPid) > 0 {
			cmd = append(cmd, "netns", nc.VethAttr.PairA.NsPid)
		}
		cmd = append(cmd, "type", "veth", "peer", "name", nc.VethAttr.PairB.Name)
		if len(nc.VethAttr.PairB.NsPid) > 0 {
			cmd = append(cmd, "netns", nc.VethAttr.PairB.NsPid)
		}
		fmt.Println(strings.Join(cmd, " "))
		err = NetRaw("", cmd...)
	} else {
		return errors.New("unknown interface type " + strconv.Itoa(int(nc.Type)))
	}
	// err := linux.Execute(context.Background(), cmd[0], cmd[1:]...).Run().Err
	if err != nil {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT)
		<-ch
		return errors.New("create interface error " + err.Error())
	}
	if nc.IsVeth() {
		// show_ip_a()
		err = IfaceUp(nc.VethAttr.PairB.Name, "")
		if err != nil {
			return errors.New("start veth pairA error " + err.Error())
		}
	}
	err = IfaceSetIP(name, ip, "", "", netns)
	if err != nil {
		return errors.New("set interface ip error " + err.Error())
	}
	if len(nc.BrdAttr.Name) > 0 {
		err = IfaceMaster(nc.VethAttr.PairB.Name, nc.BrdAttr.Name, "")
		if err != nil {
			return errors.New("set interface master error " + err.Error())
		}
	}
	//set ip4 address
	return IfaceUp(name, netns)
}
func NetRaw(netns string, args ...string) error {
	if len(args) == 0 {
		return errors.New("not set command")
	}
	var cmd []string
	if len(netns) > 0 {
		cmd = append(cmd, "ip", "netns", "exec", netns)
	}
	cmd = append(cmd, args...)
	return linux.Execute(context.Background(), cmd[0], cmd[1:]...).Run().Err
}
func IfaceUp(name string, netns string) error {
	return NetRaw(netns, "ifconfig", name, "up")
}
func IfaceDown(name string, netns string) error {
	return NetRaw(netns, "ifconfig", name, "down")
}

func IfaceSetIP(name string, addr string, netmask string, broadcast string, netns string) error {
	var cmd []string
	cmd = append(cmd, "ifconfig", name, addr)
	if len(broadcast) > 0 {
		if len(netmask) == 0 {
			netmask = "255.255.255.0"
		}
		cmd = append(cmd, "netmask", netmask, "broadcast", broadcast)
	}
	return NetRaw(netns, cmd...)
}

func IfaceMaster(ifr_name, dst_name, netns string) error {
	return NetRaw(netns, "ip", "link", "set", ifr_name, "master", dst_name)
}

func IfaceRemove(ifr_name, netns string) error {
	return NetRaw(netns, "ip", "link", "delete", ifr_name)
}

func show_ip_a() {
	linux.Execute(context.Background(), "ip", "a").Run()
}
