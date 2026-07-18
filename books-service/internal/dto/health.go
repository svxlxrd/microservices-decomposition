package dto

type HealthResponse struct {
	Status    string           `json:"status"` // "ok", "unhealthy"
	Service   string           `json:"service"`
	Version   string           `json:"version"`
	Checks    map[string]Check `json:"checks"`
	Timestamp string           `json:"timestamp"`
}

type ReadyResponse struct {
    Ready       bool              `json:"ready"`
    Service     string            `json:"service"`
    Checks      map[string]Check  `json:"checks"`
    Timestamp   string            `json:"timestamp"`
}

type Check struct {
	Status   string `json:"status"`   // "ok", "error"
	Duration string `json:"duration"` // "2ms"
	Error    string `json:"error,omitempty"`
}