package registry

type Registration struct {
	ServiceName ServiceName `json:"service_name"`
	ServiceURL  string      `json:"service_url"`
}

type ServiceName string

const (
	LogService = ServiceName("LogService")
)
