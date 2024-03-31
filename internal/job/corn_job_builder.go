package job

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
	"log"
	"strconv"
	"time"
)

type CronJobBuilder struct {
	vector *prometheus.SummaryVec
}

func NewCronJobBuilder(opt prometheus.SummaryOpts) *CronJobBuilder {
	vector := prometheus.NewSummaryVec(opt,
		[]string{"job", "success"})
	return &CronJobBuilder{
		vector: vector}

}

func (b *CronJobBuilder) Build(job Job) cron.Job {
	name := job.Name()
	return cronJobAdapterFunc(func() {
		// 接入 tracing
		start := time.Now()
		log.Println("开始运行")
		err := job.Run()
		if err != nil {
			log.Println(err)
		}
		log.Println("结束运行")
		duration := time.Since(start)
		b.vector.WithLabelValues(name, strconv.FormatBool(err == nil)).
			Observe(float64(duration.Milliseconds()))
	})
}

type cronJobAdapterFunc func()

func (c cronJobAdapterFunc) Run() {
	c()
}
