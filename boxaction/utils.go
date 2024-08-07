package boxaction

import (
	"path"
)

type _rwmode_t int

const (
	rdonly _rwmode_t = 1 << iota
	wdonly
	rw_both = rdonly | wdonly
)

func path_join(parent_path string, elem ...string) []string {
	if len(parent_path) == 0 {
		return elem
	}
	for i := range elem {
		elem[i] = path.Join(parent_path, elem[i])
	}
	return elem
}
