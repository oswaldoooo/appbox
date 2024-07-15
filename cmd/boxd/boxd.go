package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/oswaldoooo/app/cmd/boxd/boxd"
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
		err = run(&boxdcnf)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootcmd.Flags().StringP("config", "c", "", "config path")
	rootcmd.MarkFlagRequired("config")
}
func main() {
	err := rootcmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}

func run(cnf *boxd.BoxdConfig) error {
	ss, err := boxd.NewStreamService(cnf.StreamBind)
	if err != nil {
		return err
	}
	go ss.Run()

	return nil
}
