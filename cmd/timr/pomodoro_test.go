package main

import (
	"testing"
)

func TestPomodoroCommandRegistration(t *testing.T) {
	cmd := newRootCmd()

	sub, _, err := cmd.Find([]string{"pomodoro"})
	if err != nil || sub == nil {
		t.Fatal("expected pomodoro subcommand to be registered")
	}

	aliasSub, _, err := cmd.Find([]string{"p"})
	if err != nil || aliasSub == nil {
		t.Fatal("expected 'p' alias to find pomodoro subcommand")
	}

	if sub.Name() != aliasSub.Name() {
		t.Errorf("expected 'p' alias to match pomodoro command, got %s and %s", sub.Name(), aliasSub.Name())
	}
}
