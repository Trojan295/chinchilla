package gameservers

type GetLogsRequest struct {
	GameserverUUID string
	Lines          int
}

type GetLogsResponse struct {
	Logs []string
}

type LogStore interface {
	GetLogs(request *GetLogsRequest) (*GetLogsResponse, error)
}
