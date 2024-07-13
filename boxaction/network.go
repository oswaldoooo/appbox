package boxaction

import (
	"errors"
	"fmt"
	"os"

	"github.com/oswaldoooo/app/box"
	"github.com/oswaldoooo/app/internal/linux"
	"github.com/oswaldoooo/app/internal/network"
)

func BoxNetWorkBuild(cnf *box.BoxNetConfig) error {
	err := cnf.Valid()
	if err != nil {
		return err
	}
	if len(cnf.Pid) > 0 {
		err = linux.SetNsWithFile("/proc/"+cnf.Pid+"/ns/net", 0)
		if err != nil {
			return err
		}
	}
	var net_actions = map[string]func(*box.BoxNetConfig){
		"add":    add_net,
		"delete": delete_net,
		"update": update_net,
	}
	cf := net_actions[cnf.Action]
	if cf == nil {
		return errors.New("not support action " + cnf.Action)
	}
	cf(cnf)
	return nil
}

func add_net(cnf *box.BoxNetConfig) {
	err := network.NewInterface(&network.NetConfig{
		IP:       cnf.IP,
		Name:     cnf.Name,
		Type:     cnf.Type,
		BrdAttr:  cnf.BrdAttr,
		VethAttr: cnf.VethAttr,
		NsPid:    cnf.Pid,
	}, "")
	if err != nil {
		fmt.Fprintln(os.Stderr, "create interface error", err)
	}
}

func delete_net(cnf *box.BoxNetConfig) {
	err := network.IfaceRemove(cnf.Name, "")
	if err != nil {
		fmt.Fprintln(os.Stderr, "remove interface error", err)
	}
}

func update_net(cnf *box.BoxNetConfig) {

}
