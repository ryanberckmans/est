package cmd

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/ryanberckmans/est/core"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"a"},
	Short:   "Add a task",
	Long: `Add a new task

est add [(-e | --estimate) <estimate>] <task name>

The new task name is the concatenation of all non-flag args, no quotes required.
An estimate can be provided with -e, otherwise the new task will be unestimated.

// TODO move this estimate blurb to shared const between est-add and est-est
Estimates can be provided in hours "3.5h", days "2d", or weeks "0.75w". Defaults
to hours if unit unspecified. One day is equal to eight hours. One week is equal
to five days. In future, adherence to business days / hours may be configurable.

Examples:
  # Add an unestimated task named "my new task".
  est add my new task

  # Add an unestimated task named "my; new task is great".
  est add "my; new" task "is great"

  # Add an estimated task; the estimate unit defaults to hours.
  est add --estimate 2 change color to red

  # Add an estimated task; estimate is half a day; flag can go after task name.
  est add fix the bug -e 0.5d
`,
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.TrimSpace(strings.Join(args, " "))
		if len(name) < 1 {
			fmt.Println("fatal: no task name given")
			os.Exit(1)
			return
		}
		estimate, err := parseEstimate(addCmdNewTaskEstimate)
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
			ef.Tasks = append(ef.Tasks, *t)
			if err := ef.Write(); err != nil {
				fmt.Printf("fatal: %v\n", err)
				os.Exit(1)
				return
			}
			fmt.Println(core.RenderTaskOneLineSummary(t))
		}, func() {
			// failed to load estconfig or estfile. Err printed elsewhere.
		})
	},
}

var estimateRegexp = regexp.MustCompile(`^([1-9](\.[0-9]*)?$|^[0-9]\.[0-9]+)(h|d|w)?$`)

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

var addCmdNewTaskEstimate string

func init() {
	addCmd.PersistentFlags().StringVarP(&addCmdNewTaskEstimate, "estimate", "e", "", "estimate new task")
	rootCmd.AddCommand(addCmd)
}
