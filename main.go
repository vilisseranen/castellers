package main

import (
	"github.com/vilisseranen/castellers/app"
)

func main() {
	a := app.App{}
	a.Initialize()

	a.Run(":8080")
}
