package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/ryanberckmans/est/core"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"a"},
	Short:   "Add a task",
	Long: `Add a new task

est add <task name>

The new task name is the concatenation of all non-flag args, no quotes required.
An estimate can be provided with -e, otherwise the new task will be unestimated.
If an estimate was provided, the new task can be immediately started with -s.

Estimates can be provided in minutes "30m" or hours "3.5h". Estimates cannot be
provided in days or weeks, because est's auto time tracking uses customizable
working hours which makes estimation in days or weeks confusing and error-prone.
Split large tasks such that estimates are below 16 hours. See 'est schedule'.

The start time can be in the past with -a, using the same duration syntax as -e.

Examples:
  # Add an unestimated task named "my new task".
  est a my new task

  # Add an unestimated task named "my; new task is great".
  est a "my; new" task "is great"

  # Add an estimated task; the estimate unit defaults to hours.
  est a --estimate 2 change color to red

  # Add an estimated task; estimate is four hours; flag can go after task name.
  est a fix the bug -e 4h

  # Add and start an estimated task.
  est a -e 30m -s add another button

  # Add an estimated task and start it as of one hour ago.
  est a -e 4h -s -a 1h "this is a four hour task I started an hour ago"
`,
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.TrimSpace(strings.Join(args, " "))
		if len(name) < 1 {
			fmt.Println("fatal: no task name given")
			os.Exit(1)
			return
		}
		if looksLikeIDPrefix(name) {
			fmt.Println("fatal: you tried to add a task with a name that looks like an ID prefix. Did you mean another command such as 'est start'?")
			os.Exit(1)
			return
		}
		estimate, err := parseDurationHours(flagEstimate, "estimate")
		if err != nil {
			fmt.Println("fatal: " + err.Error())
			os.Exit(1)
			return
		}
		core.WithEstConfigAndFile(func(ec *core.EstConfig, ef *core.EstFile) {
			t := core.NewTask()
			if err := t.SetName(name); err != nil {
				fmt.Printf("fatal: %v\n", err)
				os.Exit(1)
				return
			}
			if estimate != 0 {
				if err := t.SetEstimated(estimate); err != nil {
					fmt.Printf("fatal: %v\n", err)
					os.Exit(1)
					return
				}
			}
			if addCmdStartNow && estimate == 0 {
				fmt.Println("fatal: cannot immediately start new task because no estimate was given")
				os.Exit(1)
				return
			}
			ef.Tasks = append(ef.Tasks, t)
			if addCmdStartNow {
				startTime := applyFlagAgo(time.Now())
				if err := ef.Tasks.Start(globalWorkTimes, len(ef.Tasks)-1, startTime); err != nil {
					fmt.Printf("fatal: %v\n", err)
					os.Exit(1)
					return
				}
				doFlagLog(t, startTime)
			} else if flagAgo != "" {
				fmt.Println("fatal: cannot add new task: -a --ago flag requires -s --start")
				os.Exit(1)
				return
			}
			if err := ef.Write(); err != nil {
				fmt.Printf("fatal: %v\n", err)
				os.Exit(1)
				return
			}
			fmt.Println(core.RenderTaskOneLineSummary(t, true))
		}, func() {
			// failed to load estconfig or estfile. Err printed elsewhere.
			os.Exit(1)
		})
	},
}

var looksLikeIDPrefixRegexp = regexp.MustCompile(`^[a-f0-9]*[0-9][a-f0-9]*$`)

// Return true iff the passed string looks like a human typed a UUID prefix.
func looksLikeIDPrefix(s string) bool {
	return looksLikeIDPrefixRegexp.MatchString(s)
}

var addCmdStartNow bool

func init() {
	addCmd.PersistentFlags().StringVarP(&flagLog, "log", "l", "", "log time worked after starting this new task")
	addCmd.PersistentFlags().StringVarP(&flagEstimate, "estimate", "e", "", "estimate new task")
	addCmd.PersistentFlags().StringVarP(&flagAgo, "ago", "a", "", "when used with start, start duration ago from now")
	addCmd.PersistentFlags().BoolVarP(&addCmdStartNow, "start", "s", false, "immediately start new task")
	rootCmd.AddCommand(addCmd)
}
