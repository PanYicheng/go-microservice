package model

type Service struct {
	Name         string
	Port         string
	Concurrency  bool
	CallList     []string
	ResponseTime float64
}

type Response struct {
	ServiceName string
	Ip          string
	Data        interface{}
	ErrorInfo   string
	Children    []Response
}

type CircuitInfo struct {
	Name                   string
	Timeout                int
	MaxConcurrentRequests  int
	SleepWindow            int
	ErrorPercentThreshold  int
	RequestVolumeThreshold int
}
