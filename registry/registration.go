package registry

type Registration struct {
	ServiceName      ServiceName   `json:"service_name"`
	ServiceURL       string        `json:"service_url"`
	RequiredServices []ServiceName `json:"required_services"`  // 需要的其他服务
	ServiceUpdateURL string        `json:"service_update_url"` // 当前服务的客户端服务
	HeartbeatURL     string        `json:"heartbeat_url"`
}

type ServiceName string

const (
	LogService  = ServiceName("LogService")
	BookService = ServiceName("BookService")
)

type patchEntry struct {
	Name ServiceName
	URL  string
}

type patch struct {
	Added   []*patchEntry
	Removed []*patchEntry
}
