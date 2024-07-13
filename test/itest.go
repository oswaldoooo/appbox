package test

import (
	"reflect"
	"testing"
)

func TestReflect(t *testing.T) {
	var tt int
	reflect.TypeOf(tt).Kind()
	reflect.ValueOf(tt)
}
