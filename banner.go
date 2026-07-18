package main

import (
	"fmt"

	"github.com/mewisme/agentrule/tui"
)

func printBanner() {
	print("\033[H\033[2J") // clear screen, cursor home
	fmt.Println(tui.BannerView())
}
