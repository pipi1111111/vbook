package ioc

import (
	rlock "github.com/gotomicro/redis-lock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
	"time"
	"vbook/internal/job"
	"vbook/internal/service"
)

func InitRankingJob(svc service.RankingService, client *rlock.Client) *job.RankingJob {
	return job.NewRankingJob(svc, time.Second*30, client)
}

func InitJobs(rjob *job.RankingJob) *cron.Cron {
	builder := job.NewCronJobBuilder(prometheus.SummaryOpts{
		Namespace: "zsz",
		Subsystem: "vbook",
		Name:      "cron_job",
		Help:      "定时任务执行",
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	})
	expr := cron.New(cron.WithSeconds())
	_, err := expr.AddJob("@every 1m", builder.Build(rjob))
	if err != nil {
		panic(err)
	}
	return expr
}
