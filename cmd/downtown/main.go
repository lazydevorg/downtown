package main

import (
	"context"
	"fmt"
)

func main() {
	app := NewApp()

	var err error
	loginRequest := LoginRequest{
		user: "fabio",
		pass: "qja.quh0ejw*ava5DEM",
	}
	loginResponse, err := app.Client.Login(context.Background(), loginRequest)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(loginResponse)

	var response Response[TasksData]
	err = app.Client.GetTasks(context.Background(), loginResponse.Data.SID, &response)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(response)
}
