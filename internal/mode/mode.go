package mode

type Mode uint8

const (
	Release Mode = 1 << iota
	Debug
	Test
)

var (
	RunMode Mode = Release
)
