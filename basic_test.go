package app_test

import (
	"path"
	"testing"
)

func TestBasic(t *testing.T) {
	t.Log(path.Join("/app", "/var/log/redis"))
}
