package main

import "fmt"

func main() {
	// err := ListTasks(true)
	// if err != nil {
	// 	fmt.Print(err)
	// }

	if err := AddNewTask("learn REST"); err != nil {
		fmt.Println(err)
	}

	if err := ListTasks(true); err != nil {
		fmt.Println(err)
	}
}
