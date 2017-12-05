package cmd

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
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

est add [(-e | --estimate) <estimate>] [--start] <task name>

The new task name is the concatenation of all non-flag args, no quotes required.
An estimate can be provided with -e, otherwise the new task will be unestimated.
If an estimate was provided, the new task can be immediately started with -s.

Estimates can be provided in minutes "30m", hours "3.5h", days "2d", or weeks
"0.75w". Defaults to hours if unit unspecified. One day is equal to eight hours.
One week is equal to five days. In future, adherence to business days / hours
may be configurable.

Examples:
  # Add an unestimated task named "my new task".
  est a my new task

  # Add an unestimated task named "my; new task is great".
  est a "my; new" task "is great"

  # Add an estimated task; the estimate unit defaults to hours.
  est a --estimate 2 change color to red

  # Add an estimated task; estimate is half a day; flag can go after task name.
  est a fix the bug -e 0.5d

  # Add and start an estimated task.
  est a -e 30m -s add another button
`,
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.TrimSpace(strings.Join(args, " "))
		if len(name) < 1 {
			fmt.Println("fatal: no task name given")
			os.Exit(1)
			return
		}
		estimate, err := parseEstimate(addCmdEstimate)
		if err != nil {
			fmt.Println("fatal: " + err.Error())
			os.Exit(1)
			return
		}
		core.WithEstConfigAndFile(func(ec *core.EstConfig, ef *core.EstFile) {
			t := core.NewTask()
			t.Name = name
			if estimate != 0 {
				t.Hours = []float64{estimate}
			}
			if addCmdStartNow && estimate == 0 {
				fmt.Println("fatal: cannot immediate start new task because no estimate was given")
				os.Exit(1)
				return
			} else if addCmdStartNow {
				t.Start(time.Now())
			}
			ef.Tasks = append(ef.Tasks, *t)
			if err := ef.Write(); err != nil {
				fmt.Printf("fatal: %v\n", err)
				os.Exit(1)
				return
			}
			fmt.Println(core.RenderTaskOneLineSummary(t))
		}, func() {
			// failed to load estconfig or estfile. Err printed elsewhere.
			os.Exit(1)
		})
	},
}

var estimateRegexp = regexp.MustCompile(`^([1-9][0-9]*(\.[0-9]*)?|0\.[0-9]+)(m|h|d|w)?$`)

// TODO move into a lib and unit test
func parseEstimate(e string) (float64, error) {
	if e == "" {
		return 0, nil
	}
	if !estimateRegexp.MatchString(e) {
		return 0, errors.New("invalid estimate")
	}
	unitMultiplier := 1.0 // default to hours
	var eWithoutUnit string
	switch e[len(e)-1:] {
	case "m":
		eWithoutUnit = e[:len(e)-1]
		unitMultiplier = 1 / 60.0 // 1/60 hours in a minute
	case "h":
		eWithoutUnit = e[:len(e)-1]
	case "d":
		eWithoutUnit = e[:len(e)-1]
		unitMultiplier = 8 // 8 hours in a day
	case "w":
		eWithoutUnit = e[:len(e)-1]
		unitMultiplier = 8 * 5 // 40 hours in a week
	default:
		eWithoutUnit = e
	}

	f, err := strconv.ParseFloat(eWithoutUnit, 64)
	if err != nil {
		return 0, errors.New("estimate wasn't a float")
	}
	return f * unitMultiplier, nil
}

var addCmdEstimate string
var addCmdStartNow bool

func init() {
	addCmd.PersistentFlags().StringVarP(&addCmdEstimate, "estimate", "e", "", "estimate new task")
	addCmd.PersistentFlags().BoolVarP(&addCmdStartNow, "start", "s", false, "immediately start new task")
	rootCmd.AddCommand(addCmd)
}
