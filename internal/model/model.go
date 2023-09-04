package model

type (
	Ping struct {
		UserId         int64
		Url            string
		ConnectionTime string
		PingTime       string
		User           User
	}

	User struct {
		Id    int64
		Login string
		Mute  bool
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
