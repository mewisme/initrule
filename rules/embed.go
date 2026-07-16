package rules

import "embed"

//go:embed *.mdc
var files embed.FS

func content(name string) ([]byte, error) {
	return files.ReadFile(name + ".mdc")
}
