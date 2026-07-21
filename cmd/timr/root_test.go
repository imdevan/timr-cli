package main

import (
	"testing"
)

func TestRootCommandHasVersion(t *testing.T) {
	cmd := newRootCmd()
	versionFlag := cmd.Flag("version")
	if versionFlag == nil {
		t.Fatal("expected --version flag to be registered")
	}
	if versionFlag.Shorthand != "V" {
		t.Fatalf("expected shorthand -V, got %q", versionFlag.Shorthand)
	}
}

func TestRootCommandHasVertical(t *testing.T) {
	cmd := newRootCmd()
	verticalFlag := cmd.Flag("vertical")
	if verticalFlag == nil {
		t.Fatal("expected --vertical flag to be registered")
	}
	if verticalFlag.Shorthand != "v" {
		t.Fatalf("expected shorthand -v, got %q", verticalFlag.Shorthand)
	}
}

func TestRootCommandHasConfig(t *testing.T) {
	cmd := newRootCmd()
	configFlag := cmd.Flag("config")
	if configFlag == nil {
		t.Fatal("expected --config flag to be registered")
	}
}

func TestResolvedVersion(t *testing.T) {
	ver := resolvedVersion()
	if ver == "" {
		t.Fatal("expected version to be non-empty")
	}
}
