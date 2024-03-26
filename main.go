package main

func main() {
	app := InitWebServer()
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}
	server := app.server
	err := server.Run(":8080")
	if err != nil {
		return
	}

}
