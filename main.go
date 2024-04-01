package main

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func main() {
	//tpCancel := ioc.InitOTEL()
	//defer func() {
	//	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//	defer cancel()
	//	tpCancel(ctx)
	//}()
	app := InitWebServer()
	initPrometheus()
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}
	//app.corn.Start()
	//defer func() {
	//	ctx := app.corn.Stop()
	//	<-ctx.Done()
	//}()
	server := app.server
	err := server.Run(":8080")
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
