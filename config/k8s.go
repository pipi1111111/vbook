//go:build k8s

package config

var Config = config{

	DB: DBConfig{
		DSN: "root:root@tcp(vbook-record-mysql:3308)/vbook",
	},
	Redis: RedisConfig{
		Addr: "vbook-record-redis:6379",
	},
}
