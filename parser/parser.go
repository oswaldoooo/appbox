package parser

import (
	"encoding/json"
	"os"
	"path"

	"github.com/oswaldoooo/app/box"
)

type ParserConfig struct {
	Err       error
	Config    box.BoxConfig
	Resources []string
}

func ParseObject(rpath string) (p ParserConfig) {
	f, err := os.OpenFile(path.Join(rpath, "config.json"), os.O_RDONLY, 0644)
	if err != nil {
		p.Err = err
		return
	}
	p.Err = json.NewDecoder(f).Decode(&p.Config)
	f.Close()
	if p.Err != nil {
		return
	}
	dents, err := os.ReadDir(rpath)
	if err != nil {
		p.Err = err
		return
	}
	p.Resources = make([]string, 0, len(dents))
	for _, d := range dents {
		if d.Name() == "config.json" {
			continue
		}
		p.Resources = append(p.Resources, d.Name())
	}
	return
}
