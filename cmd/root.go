package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/ryanberckmans/est/core"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "est",
	Short: "est is a command-line tool for software estimation",
	Long:  `est is a command-line tool for software estimation.`,
}

var globalWorkTimes core.WorkTimes

func init() {
	// TODO construct WorkTimes from estconfig
	var err error
	globalWorkTimes, err = core.New(map[time.Weekday]bool{
		time.Monday: true,
		time.Friday: true,
	}, []string{
		"8:00am",
		"10:00am",
		"12:00pm",
		"5:00pm",
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
