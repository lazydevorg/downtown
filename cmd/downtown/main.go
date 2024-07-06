package main

import (
	"fmt"
)

func main() {
	app := NewApp()

	var err error
	var tasks Response[TasksData]
	err = app.Client.GetTasks(&tasks)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(tasks.Data)
}
