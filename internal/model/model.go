package model

type (
	Ping struct {
		UserId         int64
		Url            string
		ConnectionTime string
		PingTime       string
	}

	PingList      []Ping
	TimerPingList map[string]PingList

	PingResult struct {
		Ping               Ping
		Error              error
		RealConnectionTime float64
		StatusCode         int
	}

	PingResultList []PingResult
)
