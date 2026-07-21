package domain

import (
	"testing"
)

func TestGetPomodoroPrompt(t *testing.T) {
	totalLen := 6 // Default sequence [25, 5, 25, 5, 25, 20]

	tests := []struct {
		nextIdx int
		want    string
	}{
		{0, PromptReadyToWork},
		{1, PromptReadyToBreak},
		{2, PromptReadyToWork},
		{3, PromptReadyToBreak},
		{4, PromptReadyToWork},
		{5, PromptReadyToLongBreak}, // last element is break -> Long Break
		{6, PromptReadyToGoAgain},   // after last element -> Go Again
	}

	for _, tt := range tests {
		got := GetPomodoroPrompt(tt.nextIdx, totalLen)
		if got != tt.want {
			t.Errorf("GetPomodoroPrompt(%d, %d) = %q, want %q", tt.nextIdx, totalLen, got, tt.want)
		}
	}
}

func TestGetPomodoroMessageDefaultSequence(t *testing.T) {
	msgs := DefaultPomodoroMessages()
	totalLen := 6 // Default sequence: Work(0), Break(1), Work(2), Break(3), Work(4), Break(5)

	// completedIdx = 0 (after 1st work) -> after_first_work ("You're off to a great start!" or "One small timer, one big step!")
	msg0 := GetPomodoroMessage(msgs, 0, totalLen)
	if msg0 != DefaultAfterFirstWork1 && msg0 != DefaultAfterFirstWork2 {
		t.Errorf("after 1st work got %q", msg0)
	}

	// completedIdx = 1 (after 1st break) -> after_first_break ("You got this!")
	msg1 := GetPomodoroMessage(msgs, 1, totalLen)
	if msg1 != DefaultAfterFirstBreak {
		t.Errorf("after 1st break got %q, want %q", msg1, DefaultAfterFirstBreak)
	}

	// completedIdx = 2 (after 2nd work) -> after_second_work ("Hell yeah!" or "Nice job!")
	msg2 := GetPomodoroMessage(msgs, 2, totalLen)
	if msg2 != DefaultAfterSecondWork1 && msg2 != DefaultAfterSecondWork2 {
		t.Errorf("after 2nd work got %q", msg2)
	}

	// completedIdx = 3 (after 2nd break) -> block 4 is LAST work -> before_last_work ("You're almost there!")
	msg3 := GetPomodoroMessage(msgs, 3, totalLen)
	if msg3 != DefaultBeforeLastWork {
		t.Errorf("after 2nd break got %q, want %q", msg3, DefaultBeforeLastWork)
	}

	// completedIdx = 4 (after 3rd work / LAST work) -> after_last_work ("You did it!")
	msg4 := GetPomodoroMessage(msgs, 4, totalLen)
	if msg4 != DefaultAfterLastWork {
		t.Errorf("after 3rd work got %q, want %q", msg4, DefaultAfterLastWork)
	}

	// completedIdx = 5 (after 3rd break / LAST element) -> after_last_break ("You freaking rock!")
	msg5 := GetPomodoroMessage(msgs, 5, totalLen)
	if msg5 != DefaultAfterLastBreak {
		t.Errorf("after 3rd break got %q, want %q", msg5, DefaultAfterLastBreak)
	}
}

func TestGetPomodoroMessageShortSequence(t *testing.T) {
	msgs := DefaultPomodoroMessages()
	totalLen := 2 // Work(0), Break(1)

	// completedIdx = 0 (after 1st work = LAST work) -> after_last_work ("You did it!") overrides after_first_work
	msg0 := GetPomodoroMessage(msgs, 0, totalLen)
	if msg0 != DefaultAfterLastWork {
		t.Errorf("short sequence after 1st work got %q, want %q", msg0, DefaultAfterLastWork)
	}

	// completedIdx = 1 (after 1st break = LAST element) -> after_last_break ("You freaking rock!")
	msg1 := GetPomodoroMessage(msgs, 1, totalLen)
	if msg1 != DefaultAfterLastBreak {
		t.Errorf("short sequence after 1st break got %q, want %q", msg1, DefaultAfterLastBreak)
	}
}

func TestGetPomodoroMessageLongSequence(t *testing.T) {
	msgs := DefaultPomodoroMessages()
	totalLen := 8 // Work(0), Break(1), Work(2), Break(3), Work(4), Break(5), Work(6), Break(7)

	// completedIdx = 3 (after 2nd break) -> block 4 is NOT last work (block 6 is) -> after_second_break ("")
	msg3 := GetPomodoroMessage(msgs, 3, totalLen)
	if msg3 != DefaultAfterSecondBreak {
		t.Errorf("long sequence after 2nd break got %q, want %q", msg3, DefaultAfterSecondBreak)
	}

	// completedIdx = 5 (after 3rd break) -> block 6 IS last work -> before_last_work ("You're almost there!")
	msg5 := GetPomodoroMessage(msgs, 5, totalLen)
	if msg5 != DefaultBeforeLastWork {
		t.Errorf("long sequence after 3rd break got %q, want %q", msg5, DefaultBeforeLastWork)
	}
}
