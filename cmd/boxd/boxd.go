package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"syscall"

	"github.com/emirpasic/gods/v2/maps/treemap"
	"github.com/oswaldoooo/app/cmd/boxd/boxd"
	"github.com/oswaldoooo/app/internal/unix"
	"github.com/spf13/cobra"
)

var rootcmd = cobra.Command{
	Run: func(cmd *cobra.Command, args []string) {
		cnf, _ := cmd.Flags().GetString("config")
		f, err := os.OpenFile(cnf, os.O_RDONLY, 0)
		if err != nil {
			log.Fatal(err)
		}
		var boxdcnf boxd.BoxdConfig
		err = json.NewDecoder(f).Decode(&boxdcnf)
		if err != nil {
			log.Fatal(err)
		}
		f.Close()
		err = boxdcnf.Validate()
		if err != nil {
			log.Fatal(err)
		}
		daemon_mod, _ := cmd.Flags().GetBool("daemon")
		err = run(&boxdcnf, daemon_mod)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootcmd.Flags().StringP("config", "c", "", "config path")
	rootcmd.Flags().BoolP("daemon", "d", false, "run as daemon")
	rootcmd.MarkFlagRequired("config")
}
func main() {
	err := rootcmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}

func run(cnf *boxd.BoxdConfig, daemon bool) error {
	pidmap := treemap.New[string, string]()
	ss, err := boxd.NewStreamService(cnf.StreamBind, pidmap, cnf.Logger("-stream"))
	if err != nil {
		return err
	}
	content, err := os.ReadFile("/etc/appbox/boxd.pid")
	if err == nil {
		lastpid, err := strconv.Atoi(string(content))
		if err == nil {
			syscall.Kill(lastpid, syscall.SIGKILL)
		}
	}
	if daemon {
		unix.Fork().End(-1)
	}
	go ss.Run()
	svc := boxd.NewHandService(ss, cnf.Bind, cnf.Logger("-access"))
	f, err := os.OpenFile("/etc/appbox/boxd.pid", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Fprintln(os.Stderr, "create boxd pid file error", err)
	}
	f.Write([]byte(strconv.Itoa(os.Getpid())))
	f.Close()
	return svc.Run()
}
