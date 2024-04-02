package main

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func main() {
	app := InitApp()
	initPrometheus()
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}
	err := app.server.Serve()
	if err != nil {
		return
	}
}
func initPrometheus() {
	go func() {
		//专门给 prometheus 用的端口
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(":8081", nil)
		if err != nil {
			return
		}
	}()
}
