package main

func main() {
	app := NewApp()
	web := NewWebApp(app)
	web.Start()
}
