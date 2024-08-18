package tasks

import (
	"encoding/csv"
	"fmt"
	"os"
	"slices"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/gofrs/flock" // use to lock file similar to syscall.Flock which is only for UNix
	"github.com/mergestat/timediff"
)

type Task struct {
	id         int
	name       string
	createdAt  time.Time
	isComplete bool
}

const csvFileName = "tasks.csv"

func loadFile(filePath string) (*os.File, *flock.Flock, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file for reading %w", err)
	}
	// lock the file avoid any concurrent file processes
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
	file, fileLock, err := loadFile(csvFileName)
	if err != nil {
		return nil, err
	}
	defer closeFile(file, fileLock)

	csvReader := csv.NewReader(file)

	return csvReader.ReadAll()
}

// create a new slice that holds Task type pointers with infor provided by CSV
func createAllTasks() ([]*Task, error) {
	allTasks, err := readAllTasks()
	if err != nil {
		return nil, err
	}

	taskList := make([]*Task, len(allTasks)-1)

	// loop over the task list and parse each field as per the type
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

		taskList[index] = &createTask
	}

	return taskList, nil
}

// function get the slice of Task type pointers and writes to csv file
func writeAllTasks(allTasks []*Task) error {
	taskFile, fileLock, err := loadFile(csvFileName)
	if err != nil {
		return err
	}
	defer closeFile(taskFile, fileLock)

	taskFile.Seek(0, 0)
	taskFile.Truncate(0)

	csvWriter := csv.NewWriter(taskFile)
	csvWriter.Write([]string{"ID", "Description", "CreatedAt", "IsComplete"})

	record := make([]string, 4)

	for _, task := range allTasks {
		record[0] = strconv.Itoa(task.id)
		record[1] = task.name
		record[2] = task.createdAt.Format(time.RFC3339)
		record[3] = strconv.FormatBool(task.isComplete)

		if err := csvWriter.Write(record); err != nil {
			return err
		}
	}
	csvWriter.Flush()

	return csvWriter.Error()
}

func ListTasks(listAll bool) error {
	allTasks, err := createAllTasks()
	if err != nil {
		return err
	}

	if len(allTasks) < 1 {
		return fmt.Errorf("no tasks found, add new tasks")
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)

	if !listAll {
		fmt.Fprint(w, "ID\tTask\tCreated\n")

		for _, task := range allTasks {
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

		for _, task := range allTasks {
			fmt.Fprintf(w, "%d\t%s\t%s\t%v\n", task.id, task.name, timediff.TimeDiff(task.createdAt), task.isComplete)
		}
		w.Flush()
	}

	return nil
}

// get the last id from the task which can be incremented on each new task
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

	nowStr := time.Now().Format(time.RFC3339)

	file, fileLock, err := loadFile(csvFileName)
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

func CompleteTask(taskId int) error {
	allTasks, err := createAllTasks()
	if err != nil {
		return err
	}

	idExists := false
	for _, task := range allTasks {
		if task.id == taskId {
			task.isComplete = true
			idExists = true
			break
		}
	}

	if !idExists {
		return fmt.Errorf("no task with given id")
	}

	return writeAllTasks(allTasks)
}

func DeleteTask(taskId int) error {
	allTasks, err := createAllTasks()
	if err != nil {
		return err
	}

	var deletIndex int
	idExists := false
	for i, task := range allTasks {
		if task.id == taskId {
			deletIndex = i
			idExists = true
			break
		}
	}

	if !idExists {
		return fmt.Errorf("no task with given id")
	}

	allTasks = slices.Delete(allTasks, deletIndex, deletIndex+1)
	return writeAllTasks(allTasks)
}
