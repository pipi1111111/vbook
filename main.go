package main

func main() {
	server := InitWebServer()
	err := server.Run(":8080")
	if err != nil {
		return
	}

}
