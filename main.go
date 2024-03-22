package main

import "vbook/internal/integration/startup"

func main() {
	server := startup.InitWebServer()
	err := server.Run(":8080")
	if err != nil {
		return
	}

}
