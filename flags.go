package main

import (
	"github.com/LimeHD/limehd-syslog-server/constants"
	"github.com/urfave/cli"
)

var CliFlags = []cli.Flag{
	&cli.BoolFlag{
		Name:  "debug",
		Usage: "Режим отладки - детализирует этапы работы сервиса, также в данном режиме все логи отправляются в stdout",
	},
	&cli.StringFlag{
		Name:     "bind-address",
		Usage:    "IP и порт слушаетля syslog, например: 0.0.0.0:514",
		Required: true,
	},
	&cli.StringFlag{
		Name:     "log",
		Usage:    "Файл, куда будут складываться логи",
		Required: false,
	},
	&cli.StringFlag{
		Name:  "maxmind",
		Usage: "Файл базы данных MaxMind с расширением .mmdb",
		Value: constants.DEFAULT_MAXMIND_DATABASE,
	},
	&cli.StringFlag{
		Name:  "maxmind-asn",
		Usage: "Файл базы данных MaxMind ASN с расширением .mmdb для автономных систем",
		Value: constants.DEFAULT_MAXMIND_ASN_DATABASE,
	},
	&cli.StringFlag{
		Name:  "statsd-address",
		Usage: "Ip адрес и порт подключения к statsd, например: 127.0.0.1:8125",
		Value: "127.0.0.1:8125",
	},
	&cli.StringFlag{
		Name:     "statsd-measurement",
		Usage:    "Название измерения (measurement) в statsd",
		Required: true,
	},
	&cli.StringFlag{
		Name:     "statsd-measurement-online",
		Usage:    "Название измерения (measurement) в statsd для счетчиков online пользователей",
		Required: true,
	},
	&cli.Int64Flag{
		Name:     "online-duration",
		Usage:    "За какой промежуток агрегировать уникальных пользователей (в секундах)",
		Value:    300,
		Required: true,
	},
	&cli.StringFlag{
		Name:  "nginx-template",
		Usage: "Шаблон для конфигурации форматов логов Nginx",
		Value: "./template.conf",
	},
	&cli.IntFlag{
		Name:  "pool-size",
		Usage: "Максимальная емкость воркеров для обработки запросов",
		Value: 4000000,
	},
	&cli.IntFlag{
		Name:  "worker-pool-size",
		Usage: "Максимальная емкость воркеров для обработки запросов",
		Value: 2000000,
	},
	&cli.IntFlag{
		Name:  "max-parallel",
		Usage: "Максимальное количество параллельных обработчиков для входящих UPD запросов",
		Value: 15,
	},
	&cli.IntFlag{
		Name:  "worker-count",
		Usage: "Максимальное количество обработчиков логов",
		Value: 35,
	},
	&cli.IntFlag{
		Name:  "sender-count",
		Usage: "Максимальное количество и отправителей в Influx",
		Value: 10,
	},
}
