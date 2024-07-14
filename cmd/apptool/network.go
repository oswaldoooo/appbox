package main

import (
	"errors"

	"github.com/oswaldoooo/app/box"
	"github.com/oswaldoooo/app/boxaction"
	"github.com/oswaldoooo/app/internal/network"
)

func network_init(pid string, cnf box.BoxNetConfig) error {
	cnf.Action = "add"
	network.TargetNs = pid
	cnf.Pid = pid
	println("pid", pid)
	err := network.Lock()
	if err != nil {
		return errors.New("lock error " + err.Error())
	}
	defer network.Unlock()
	err = boxaction.BoxNetWorkBuild(&cnf)
	if err != nil {
		return errors.New("build network error " + err.Error())
	}
	var nc network.NetConfig
	nc.SetBrd(cnf.Name, cnf.IP, cnf.KeepBit, pid)
	err = network.Dump(&nc)
	if err != nil {
		return errors.New("dump network error " + err.Error())
	}
	return nil
}
