package main

import (
	"github.com/vilisseranen/castellers/app"
)

func main() {
	a := app.App{}

	otelClose := a.Initialize()
	defer otelClose()

	a.Run(":8080")
}
