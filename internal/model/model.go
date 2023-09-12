package model

type (
	// Ping моделька для представления записи в тиблице ping
	Ping struct {
		UserId         int64
		Url            string
		ConnectionTime string
		PingTime       string
		User           User
	}

	// User моделька для представления записи в тиблице users
	User struct {
		Id    int64
		Login string
		Mute  bool
	}

	// PingList моделька для представления списка записей из таблици ping
	PingList []Ping
	// TimerPingList отформатированый список записей с таблици ping по времени опроса: {"30s":PingList, "40s":PingList}
	TimerPingList map[string]PingList

	// PingResult моделька содержит информацию о результате опроса ссылки из таблици ping
	PingResult struct {
		Ping               Ping
		Error              error
		RealConnectionTime float64
		StatusCode         int
		IsCancel           bool
	}

	PingResultList []PingResult // see PingResult

	// Statistic данные по пингам
	Statistic struct {
		Url               string
		CountPing         int
		CorrectCount      int
		CancelCount       int
		MaxConnectionTime float64
		MinConnectionTime float64
		AvgConnectionTime float64
	}

	StatisticResultList []Statistic
)
