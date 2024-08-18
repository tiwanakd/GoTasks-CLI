package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tiwanakd/GoTasks-CLI.git/tasks"
)

var (
	rootCmd = &cobra.Command{Use: "tasks"}

	listAll      bool
	cmdListTasks = &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := tasks.ListTasks(listAll); err != nil {
				return err
			}
			return nil
		},
	}
	cmdAddNewTask = &cobra.Command{
		Use:  "add [name of task]",
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			tasks.AddNewTask(strings.Join(args, " "))
		},
	}
	cmdCompleteTask = &cobra.Command{
		Use:  "complete [taskId]",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("invalid number of arguments provided")
			}
			if err := tasks.CompleteTask(convertID(args)); err != nil {
				return err
			}
			return nil
		},
	}
	cmdDeleteTask = &cobra.Command{
		Use:  "delete [taskId]",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("invalid number of arguments provided")
			}
			if err := tasks.DeleteTask(convertID(args)); err != nil {
				return err
			}
			return nil
		},
	}
)

func convertID(args []string) int {
	var taskId int
	var err error
	for _, id := range args {
		taskId, err = strconv.Atoi(id)
		if err != nil {
			return -1
		}
		break
	}
	return taskId
}

func init() {
	cmdListTasks.Flags().
		BoolVarP(&listAll, "all", "a", false, "list all completed and uncompleted tasks")

	rootCmd.AddCommand(cmdListTasks)
	rootCmd.AddCommand(cmdAddNewTask)
	rootCmd.AddCommand(cmdCompleteTask)
	rootCmd.AddCommand(cmdDeleteTask)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
