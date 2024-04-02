package main

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
	"net/http"
)

func main() {
	initViperWatch()
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
func initViperWatch() {
	cfile := pflag.String("config",
		"config/config.yaml", "配置文件路径")
	// 这一步之后，cfile 里面才有值
	pflag.Parse()
	//viper.Set("db.dsn", "localhost:3306")
	// 所有的默认值放好s
	viper.SetConfigType("yaml")
	viper.SetConfigFile(*cfile)
	viper.WatchConfig()
	// 读取配置
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	val := viper.Get("test.key")
	log.Println(val)
}
