package main

import (
	"os"
)

func main() {
	a := App{}
	a.Initialize(os.Getenv("APP_DB_NAME"), os.Getenv("APP_LOG_FILE"))

	a.Run(":8080")
}
