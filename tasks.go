package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/gofrs/flock"
	"github.com/mergestat/timediff"
)

type Task struct {
	id         int
	name       string
	createdAt  time.Time
	isComplete bool
}

func loadFile(filePath string) (*os.File, *flock.Flock, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file for reading %w", err)
	}

	fileLock := flock.New(filePath)
	locked, err := fileLock.TryLock()
	if err != nil {
		file.Close()
		return nil, nil, fmt.Errorf("error locking the filen %w", err)
	}

	if locked {
		return file, fileLock, nil
	} else {
		file.Close()
		return nil, nil, fmt.Errorf("unable to lock file %s", filePath)
	}
}

func closeFile(file *os.File, fileLock *flock.Flock) error {
	fileLock.Unlock()
	return file.Close()
}

func readAllTasks() ([][]string, error) {
	fileName := "tasks.csv"

	file, fileLock, err := loadFile(fileName)
	if err != nil {
		return nil, err
	}
	defer closeFile(file, fileLock)

	csvReader := csv.NewReader(file)

	return csvReader.ReadAll()
}

func createAllTasks() (*[]Task, error) {
	allTasks, err := readAllTasks()
	if err != nil {
		return nil, err
	}

	taskList := make([]Task, len(allTasks)-1)

	for index, task := range allTasks[1:] {
		taskId, err := strconv.Atoi(task[0])
		if err != nil {
			return nil, err
		}

		isComplete, err := strconv.ParseBool(task[3])
		if err != nil {
			return nil, err
		}

		createdAt, err := time.Parse(time.RFC3339, task[2])
		if err != nil {
			return nil, err
		}

		createTask := Task{
			id:         taskId,
			name:       task[1],
			createdAt:  createdAt,
			isComplete: isComplete,
		}

		taskList[index] = createTask
	}

	return &taskList, nil
}

func ListTasks(listAll bool) error {
	allTasks, err := createAllTasks()
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)

	if !listAll {
		fmt.Fprint(w, "ID\tTask\tCreated\n")

		for _, task := range *allTasks {
			if !task.isComplete {
				fmt.Fprintf(
					w,
					"%d\t%s\t%s\n",
					task.id,
					task.name,
					timediff.TimeDiff(task.createdAt),
				)
			}
		}
		w.Flush()
	} else {
		fmt.Fprint(w, "ID\tTask\tCreated\tDone\n")

		for _, task := range *allTasks {
			fmt.Fprintf(w, "%d\t%s\t%s\t%v\n", task.id, task.name, timediff.TimeDiff(task.createdAt), task.isComplete)
		}
		w.Flush()
	}

	return nil
}

func getLastTaskID() (int, error) {
	allTasks, err := readAllTasks()
	if err != nil {
		return -1, err
	}

	if len(allTasks) < 2 {
		return 0, nil
	}
	lastTask := allTasks[len(allTasks)-1]

	return strconv.Atoi(lastTask[0])
}

func AddNewTask(name string) error {
	lastTaskId, err := getLastTaskID()
	if err != nil {
		return err
	}

	newTaskId := strconv.Itoa(lastTaskId + 1)

	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	file, fileLock, err := loadFile("tasks.csv")
	if err != nil {
		return err
	}
	defer closeFile(file, fileLock)

	newTaskRecord := []string{
		newTaskId, name, nowStr, "false",
	}
	csvWriter := csv.NewWriter(file)
	if err := csvWriter.Write(newTaskRecord); err != nil {
		return err
	}
	csvWriter.Flush()

	return csvWriter.Error()
}
