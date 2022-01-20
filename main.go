package main

import (
	"catya/api"
	"catya/app"
)

func main() {
	huya := api.New()
	application := app.New(&huya)
	application.Run()
}
