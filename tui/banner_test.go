package tui

import (
	"strings"
	"testing"
)

func TestBannerViewNonEmpty(t *testing.T) {
	s := BannerView()
	if s == "" {
		t.Fatal("empty banner")
	}
	if strings.Count(s, "\n") < 3 {
		t.Fatal("banner too short")
	}
	if BannerHeight() < 1 {
		t.Fatal("banner height")
	}
}

func TestRenderBannerNarrow(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic: %v", r)
		}
	}()
	_ = RenderBanner(1)
	_ = RenderBanner(0)
}
