package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/oswaldoooo/app/box"
	"github.com/oswaldoooo/app/internal/mode"
	"github.com/oswaldoooo/app/internal/network"
	"github.com/spf13/cobra"
)

func init() {
	appmode := strings.ToLower(os.Getenv("APP_MODE"))
	if appmode == "debug" {
		mode.RunMode |= mode.Debug
	}
}
func main() {
	var rootcmd, networkcmd cobra.Command
	networkcmd.Use = "network"
	networkcmd.Short = "network controller"
	networkcmd.AddCommand(NewNetworkInitCommand())
	rootcmd.AddCommand(&networkcmd)
	err := rootcmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error", err)
		os.Exit(-1)
	}
}

func NewNetworkInitCommand() *cobra.Command {
	var (
		cnf    box.BoxNetConfig
		set_ip box.IP
	)
	var cmd = cobra.Command{
		Use:   "init",
		Short: "init namespace network",
		Run: func(cmd *cobra.Command, args []string) {
			cnf.Type = network.Brd
			cnf.IP = set_ip.IP
			cnf.KeepBit = set_ip.KeepBit
			cnf.Name = args[0]
			_pid, _ := cmd.Flags().GetString("pid")
			err := network_init(_pid, cnf)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
	cmd.Flags().String("pid", "", "target namepsace pid")
	cmd.Flags().Var(&set_ip, "ip", "init bridge ip")
	return &cmd
}
