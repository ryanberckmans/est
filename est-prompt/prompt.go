package main

import (
	"fmt"
	"strings"

	"github.com/ryanberckmans/est/core"
)

const ansiReset = "\033[0m"
const ansiBold = "\033[1m"
const ansiBoldYellow = "\033[93m"
const ansiBoldMagenta = "\033[95m"
const ansiDangerOrange = "\033[38;5;202m" // color 202 of 256

// For bash, escape the colors with \[ \] http://mywiki.wooledge.org/BashFAQ/053
// . The \[ \] are only special when you assign PS1, if you print
// them inside a function that runs when the prompt is displayed it
// doesn't work. In this case you need to use the bytes \001 and \002:
const bashOpen = "\001"
const bashClose = "\002"

const promptNoTasksStarted = bashOpen + ansiDangerOrange + bashClose + "no tasks started      " + bashOpen + ansiReset + bashClose

// When the prompt cannot be displayed, provide a default prompt
// which will cause minimal disruption to the user's prompt.
const promptFailed = bashOpen + ansiDangerOrange + bashClose + "est-prompt failed     " + bashOpen + ansiReset + bashClose

// renderPrompt renders a summary of passed tasks, such that returned string is
// suitable to be used as part of a shell prompt.
// The rendering aims to be minimally distracting by being fixed width, commas always
// in same char position, and same color; and maximally useful, currently by showing
// an adaptive short form of task names.
func renderPrompt(ts []*core.Task) string {
	switch len(ts) {
	case 0:
		return promptNoTasksStarted
	case 1:
		return promptColor(promptOneTask(ts[0]))
	case 2:
		return promptColor(promptTwoTasks(ts[0], ts[1]))
	default:
		return promptColor(promptNTasks(ts[0], ts[1], ts[2], ts[3:]))
	}
}

func promptColor(s string) string {
	return bashOpen + ansiReset + ansiBold + ansiBoldYellow + bashClose + s + bashOpen + ansiReset + bashClose
}

func promptOneTask(t *core.Task) string {
	s := getTaskNameForPrompt(t)
	if len(s) > 22 {
		return s[:22]
	}
	return fmt.Sprintf("%-22s", s)
}

func promptTwoTasks(t *core.Task, t2 *core.Task) string {
	s := stringsShortForm(5, 10, strings.Fields(getTaskNameForPrompt(t)))
	s2 := stringsShortForm(5, 10, strings.Fields(getTaskNameForPrompt(t2)))
	return fmt.Sprintf("%10s, %-10s", s, s2)
}

// fixed width relies on len(ts) < 9, i.e. no more than 12 tasks in progress.
func promptNTasks(t *core.Task, t2 *core.Task, _ *core.Task, ts []*core.Task) string {
	s := stringsShortForm(5, 10, strings.Fields(getTaskNameForPrompt(t)))
	s2 := stringsShortForm(4, 8, strings.Fields(getTaskNameForPrompt(t2)))
	return fmt.Sprintf("%10s, %-8s+%d", s, s2, 1+len(ts))
}

func getTaskNameForPrompt(t *core.Task) string {
	// TODO t.ShortName is used if non-empty; maybe t.Summary instead; or t.Tags
	return t.Name()
}

// stringsShortForm returns a short form for passed string slice
// which won't exceed passed maxLen, earlier elements of slice
// won't be truncated below passed minTokenLen.
func stringsShortForm(minTokenLen int, maxLen int, ss []string) string {
	if maxLen < 1 {
		return ""
	}
	switch len(ss) {
	case 0:
		panic("stringsShortForm: unexpected empty string slice")
	case 1:
		// single token remains, return longest prefix within maxLen
		headLen := minInt(len(ss[0]), maxLen)
		return ss[0][:headLen]
	}
	// ss has multiple tokens. Get short form of tail first, so that
	// we know how much remaining length to add to head.
	tailMaxLen := maxLen - minInt(minTokenLen, len(ss[0])) - 1 // the max length for tail tokens is current maxLen, minus one for space between head and next token, minus the min length of head token. In this way, extra space is first passed down to tail tokens (when head token < minTokenLen), and then extra space is passed back up to head token (when tail tokens < their maxLen).

	if tailMaxLen < 1 {
		// no length left for tail, return largest head
		headLen := minInt(len(ss[0]), maxLen)
		return ss[0][:headLen]
	}

	tailShortForm := stringsShortForm(minTokenLen, tailMaxLen, ss[1:])
	headLen := minInt(len(ss[0]), maxLen-len(tailShortForm)-1)
	return ss[0][:headLen] + " " + tailShortForm
}

func minInt(i, j int) int {
	if i < j {
		return i
	}
	return j
}
