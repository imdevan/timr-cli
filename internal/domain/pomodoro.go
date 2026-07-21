package domain

import (
	"math/rand"
)

const (
	PromptReadyToWork      = "Ready to work?"
	PromptReadyToBreak     = "Ready to take a break?"
	PromptReadyToLongBreak = "Ready to take a long break?"
	PromptReadyToGoAgain   = "Ready to go again?"

	DefaultAfterFirstWork1  = "You're off to a great start!"
	DefaultAfterFirstWork2  = "One small timer, one big step!"
	DefaultAfterFirstBreak  = "You got this!"
	DefaultAfterSecondWork1 = "Hell yeah!"
	DefaultAfterSecondWork2 = "Nice job!"
	DefaultAfterSecondBreak = ""
	DefaultBeforeLastWork   = "You're almost there!"
	DefaultAfterLastWork    = "You did it!"
	DefaultAfterLastBreak   = "You freaking rock!"
)

// PomodoroMessagesConfig holds user-configurable message sets for pomodoro phase transitions.
type PomodoroMessagesConfig struct {
	AfterFirstWork   []string `toml:"after_first_work"`
	AfterFirstBreak  []string `toml:"after_first_break"`
	AfterSecondWork  []string `toml:"after_second_work"`
	AfterSecondBreak []string `toml:"after_second_break"`
	BeforeLastWork   []string `toml:"before_last_work"`
	AfterLastWork    []string `toml:"after_last_work"`
	AfterLastBreak   []string `toml:"after_last_break"`
}

// DefaultPomodoroMessages returns default pomodoro transition messages.
func DefaultPomodoroMessages() PomodoroMessagesConfig {
	return PomodoroMessagesConfig{
		AfterFirstWork:   []string{DefaultAfterFirstWork1, DefaultAfterFirstWork2},
		AfterFirstBreak:  []string{DefaultAfterFirstBreak},
		AfterSecondWork:  []string{DefaultAfterSecondWork1, DefaultAfterSecondWork2},
		AfterSecondBreak: []string{DefaultAfterSecondBreak},
		BeforeLastWork:   []string{DefaultBeforeLastWork},
		AfterLastWork:    []string{DefaultAfterLastWork},
		AfterLastBreak:   []string{DefaultAfterLastBreak},
	}
}

// PickRandomString selects a random string from a list, returning "" if empty.
func PickRandomString(list []string) string {
	if len(list) == 0 {
		return ""
	}
	if len(list) == 1 {
		return list[0]
	}
	return list[rand.Intn(len(list))]
}

// GetPomodoroMessage calculates the contextual transition message after block completedIdx in a sequence of length totalLen.
// completedIdx is 0-indexed:
// 0: after 1st block (1st work)
// 1: after 2nd block (1st break)
// ...
func GetPomodoroMessage(msgs PomodoroMessagesConfig, completedIdx int, totalLen int) string {
	if totalLen <= 0 || completedIdx < 0 || completedIdx >= totalLen {
		return ""
	}

	// Calculate indices of work and break blocks (assuming 0 is 1st work, 1 is 1st break, etc.)
	var workIndices []int
	var breakIndices []int
	for i := 0; i < totalLen; i++ {
		if i%2 == 0 {
			workIndices = append(workIndices, i)
		} else {
			breakIndices = append(breakIndices, i)
		}
	}

	lastWorkIdx := -1
	if len(workIndices) > 0 {
		lastWorkIdx = workIndices[len(workIndices)-1]
	}

	// Rule: after last element overall
	if completedIdx == totalLen-1 {
		if len(breakIndices) > 0 && completedIdx == breakIndices[len(breakIndices)-1] {
			return PickRandomString(msgs.AfterLastBreak)
		}
		if completedIdx == lastWorkIdx {
			return PickRandomString(msgs.AfterLastWork)
		}
	}

	// Rule: completed block IS the last work block
	if completedIdx == lastWorkIdx {
		return PickRandomString(msgs.AfterLastWork)
	}

	// Rule: NEXT block (completedIdx + 1) IS the last work block
	if completedIdx+1 == lastWorkIdx {
		return PickRandomString(msgs.BeforeLastWork)
	}

	// Positional defaults
	switch completedIdx {
	case 0:
		return PickRandomString(msgs.AfterFirstWork)
	case 1:
		return PickRandomString(msgs.AfterFirstBreak)
	case 2:
		return PickRandomString(msgs.AfterSecondWork)
	case 3:
		return PickRandomString(msgs.AfterSecondBreak)
	default:
		return ""
	}
}

// GetPomodoroPrompt determines the prompt string depending on the next block to start (nextIdx).
// nextIdx is 0-indexed:
// 0: before 1st block
// 1: before 2nd block
// ...
// totalLen: after last element overall
func GetPomodoroPrompt(nextIdx int, totalLen int) string {
	if nextIdx >= totalLen {
		return PromptReadyToGoAgain
	}

	// Check if nextIdx is a work or break block
	isWork := (nextIdx % 2 == 0)
	if isWork {
		return PromptReadyToWork
	}

	// Break block
	if nextIdx == totalLen-1 {
		return PromptReadyToLongBreak
	}
	return PromptReadyToBreak
}
