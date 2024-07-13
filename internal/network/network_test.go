package network_test

import (
	"testing"

	"github.com/oswaldoooo/app/internal/network"
)

func TestGetConfig(t *testing.T) {
	cnf, err := network.GetNetConfig()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(cnf)
	cnf, err = network.NewSubnet(&cnf)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(cnf)
}
