package rules

import "embed"

//go:embed rewrite/*
var files embed.FS

func content(name string) ([]byte, error) {
	return files.ReadFile("rewrite/" + name + ".mdc")
}
