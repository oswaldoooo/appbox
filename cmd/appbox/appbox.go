package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/oswaldoooo/app/box"
	"github.com/oswaldoooo/app/boxaction"
	"github.com/oswaldoooo/app/internal/linux"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func main() {
	var rootcmd cobra.Command
	rootcmd.AddCommand(NewRunCommand(), NewListCommand(), NewNetCommand(), NewNsExecCommand())
	err := rootcmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, "programer error", err)
		os.Exit(-1)
	}
}

func NewRunCommand() *cobra.Command {
	var bcnf box.BoxConfig
	var cmds = cobra.Command{
		Use:           "run",
		Short:         "run",
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			rpath, _ := cmd.Flags().GetString("apply")
			if len(rpath) > 0 {
				content, err := os.ReadFile(rpath)
				throw(err, "open "+rpath+" error")
				err = json.Unmarshal(content, &bcnf)
				throw(err, "format error")
			}
			throw(bcnf.Valid(), "config error")
			err := boxaction.BoxBuild(context.Background(), bcnf)
			throw(err, "box build error")

		},
	}
	cmds.Flags().StringP("apply", "f", "", "apply config file")
	cmds.Flags().StringSliceVar(&bcnf.Path, "path", []string{}, "path")
	cmds.Flags().StringSliceVar(&bcnf.StaticPath, "static-path", []string{}, "static path")
	cmds.Flags().BoolVarP(&bcnf.StandloneNet, "net", "n", false, "stand lone network")
	cmds.Flags().BoolVar(&bcnf.StandloneFs, "fs", false, "stand lone fs")
	cmds.Flags().BoolVarP(&bcnf.StandloneUsers, "user", "u", false, "stand lone user")
	cmds.Flags().BoolVar(&bcnf.StandloneHost, "host", false, "stand lone host")
	cmds.Flags().BoolVar(&bcnf.LinkMode, "link-mode", false, "link mode")
	cmds.Flags().StringVar(&bcnf.LinkFs, "link-fs", "", "link fs pid")
	cmds.Flags().StringVar(&bcnf.LinkNet, "link-net", "", "link net pid")
	cmds.Flags().StringVar(&bcnf.LinkUts, "link-uts", "", "link uts pid")
	cmds.Flags().StringVar(&bcnf.LinkPid, "link-pid", "", "link pid namespace pid")
	return &cmds
}

func NewListCommand() *cobra.Command {

	var cmd = cobra.Command{
		Use:           "list",
		Short:         "show appbox list",
		SilenceErrors: true,
		SilenceUsage:  true,
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
	return &cmd
}

var netcnf box.BoxNetConfig

func NewNetCommand() *cobra.Command {

	var cmd = cobra.Command{
		Use:   "network",
		Short: "appbox network manage",
	}
	var cmds = []*cobra.Command{&cobra.Command{
		Use:           "add",
		Short:         "add appbox network config",
		SilenceErrors: true,
		SilenceUsage:  true,
		Run:           do_net,
	}, &cobra.Command{
		Use:           "delete",
		Short:         "delete appbox network config",
		SilenceErrors: true,
		SilenceUsage:  true,
		Run:           do_net,
	}}
	do_net_flag(cmds[0].Flags(), &netcnf)
	do_net_flag(cmds[1].Flags(), &netcnf)
	cmd.AddCommand(cmds...)
	return &cmd
}
func do_net_flag(pf *pflag.FlagSet, bcnf *box.BoxNetConfig) {
	pf.StringVar(&bcnf.Name, "name", "", "interface name")
	pf.IPVar(&bcnf.IP, "ip", net.IP{}, "interface ip")
	pf.StringVar(&bcnf.BrdAttr.Name, "gateway-name", "", "veth bind bridge interface name")
	pf.StringVar(&bcnf.Pid, "pid", "", "target pid[optional]")
	pf.Var(&bcnf.Type, "type", "net type")
	pf.StringP("apply", "f", "", "apply config file")
	pf.Uint8Var(&bcnf.KeepBit, "keep", 0, "net keep")
}
func do_net(cmd *cobra.Command, args []string) {
	fpath, _ := cmd.Flags().GetString("apply")
	if len(fpath) > 0 {
		err := loadobj(fpath, &netcnf)
		if err != nil {
			fmt.Fprintln(os.Stderr, "read config", fpath, "error", err)
			return
		}
	}
	netcnf.Action = cmd.Use
	err := boxaction.BoxNetWorkBuild(&netcnf)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error", err)
	}
}
func throw(err error, words string) {
	if err != nil {
		fmt.Fprintln(os.Stderr, words, err)
		os.Exit(-1)
	}
}

func loadobj(rpath string, v any) error {
	f, err := os.OpenFile(rpath, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(v)
}

func NewNsExecCommand() *cobra.Command {
	type NsExecFlag struct {
		Net    bool
		Ns     bool
		Pid    bool
		Uts    bool
		Target string
	}
	var ne_flag NsExecFlag
	var cmd = cobra.Command{
		Use:   "nsexec",
		Short: "exec command in target namespace",
		Run: func(cmd *cobra.Command, args []string) {
			_flag := 0
			if ne_flag.Net {
				_flag |= linux.CLONE_NET
			}
			if ne_flag.Ns {
				_flag |= linux.CLONE_NS
			}
			if ne_flag.Pid {
				_flag |= linux.CLONE_PID
			}
			if ne_flag.Uts {
				_flag |= linux.CLONE_UTS
			}
			parent_pid := os.Getpid()
			linux.NsExec(_flag, ne_flag.Target).Then(func(h *linux.Hook, a any) {
				_arg := a.([]string)
				err := linux.Execute(context.Background(), _arg[0], _arg[1:]...).SetIO(os.Stdout, os.Stderr).SetStdin(os.Stdin).Run().Err
				if err != nil {
					fmt.Fprintln(os.Stderr, "execute command error", err)
				}
				syscall.Kill(parent_pid, syscall.SIGINT)
			}, args).End(-1)
			ch := make(chan os.Signal, 1)
			signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
			<-ch
		},
	}
	cmd.Flags().StringVar(&ne_flag.Target, "target", "", "target pid namespace")
	cmd.Flags().BoolVarP(&ne_flag.Net, "net", "n", false, "net namespace")
	cmd.Flags().BoolVarP(&ne_flag.Ns, "mnt", "m", false, "mount namespace")
	cmd.Flags().BoolVarP(&ne_flag.Pid, "pid", "p", false, "pid namespace")
	cmd.Flags().BoolVarP(&ne_flag.Uts, "uts", "u", false, "uts namespace")
	cmd.MarkFlagRequired("target")
	return &cmd
}
