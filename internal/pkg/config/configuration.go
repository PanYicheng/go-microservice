package config

type Config struct {
	Environment     string `arg:"env:ENVIRONMENT" default:"dev"`
	ServicePort string `arg:"env:SERVICE_PORT" default:"6767"`
	ServiceName string `arg:"env:SERVICE_NAME" default:"accountservice"`
	ZipkinUrl string `arg:"env:ZIPKIN_URL" default:"http://zipkin:9411"`
	AmqpUrl string `arg:"env:AMQP_URL" default:"amqp://guest:guest@rabbitmq:5672/"`
}
