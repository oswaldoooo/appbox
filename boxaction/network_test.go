package boxaction_test

import (
	"net"
	"testing"

	"github.com/oswaldoooo/app/box"
	"github.com/oswaldoooo/app/boxaction"
	"github.com/oswaldoooo/app/internal/network"
)

func TestBridgeCreate(t *testing.T) {
	err := boxaction.BoxNetWorkBuild(&box.BoxNetConfig{
		Action: "add",
		IP:     net.IPv4(172, 20, 0, 1),
		Name:   "br0",
		Type:   network.Brd,
		Pid:    "86674",
	})
	if err != nil {
		t.Fatal("box network build error", err)
	}
}

func TestIfaceDelete(t *testing.T) {
	err := boxaction.BoxNetWorkBuild(&box.BoxNetConfig{
		Action: "delete",
		Name:   "br0",
		Pid:    "86674",
	})
	if err != nil {
		t.Fatal(err)
	}
}
