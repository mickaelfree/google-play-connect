package cmd_test

import (
	"testing"

	"github.com/mickaelfree/google-play-connect/internal/cmd"
)

func TestRequireConfirmFlagSet(t *testing.T) {
	if err := cmd.RequireConfirm(true, func(string) string { return "" }, "publish"); err != nil {
		t.Fatalf("confirm=true should pass: %v", err)
	}
}

func TestRequireConfirmCI(t *testing.T) {
	getenv := func(k string) string {
		if k == "CI" {
			return "true"
		}
		return ""
	}
	if err := cmd.RequireConfirm(false, getenv, "publish"); err != nil {
		t.Fatalf("CI=true should pass: %v", err)
	}
}

func TestRequireConfirmRefuses(t *testing.T) {
	err := cmd.RequireConfirm(false, func(string) string { return "" }, "publish release")
	if err == nil {
		t.Fatal("expected error without --confirm")
	}
}
