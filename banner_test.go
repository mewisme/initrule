package main

import (
	"strings"
	"testing"

	"github.com/mewisme/agentrule/tui"
)

func TestBannerViewSharedByPrintBanner(t *testing.T) {
	// printBanner prints tui.BannerView(); assert the shared source is non-empty.
	if !strings.Contains(tui.BannerView(), "█") {
		t.Fatal("BannerView empty")
	}
}
