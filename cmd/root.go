package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/ryanberckmans/est/core/worktimes"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "est",
	Short: "est is a command-line tool for software estimation",
	Long:  `est is a command-line tool for software estimation.`,
}

var globalWorkTimes worktimes.WorkTimes

func init() {
	// TODO construct WorkTimes from estconfig
	var err error
	globalWorkTimes, err = worktimes.New(map[time.Weekday]bool{
		time.Monday:    true,
		time.Tuesday:   true,
		time.Wednesday: true,
		time.Thursday:  true,
		time.Friday:    true,
	}, []string{
		// Work 9:30am-noon
		"9:30am",
		"12:00pm",
		// 30 minutes for lunch, then work 12:30pm-5:30pm
		"12:30pm",
		"5:30pm",
		// Time on tasks outside of these times will not count towards automatic time tracking. This doesn't mean no work occurs outside of these times, it just means the estimator isn't penalized (by additional auto time tracking duration) for not making progress during non-working hours. `est log` can also be used as an escape hatch, e.g. for a 3h project on a Saturday.
	})
	if err != nil {
		fmt.Println("fatal: " + err.Error())
		os.Exit(1)
		return
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
