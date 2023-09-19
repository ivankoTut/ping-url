package model

type (
	// Ping моделька для представления записи в тиблице ping
	Ping struct {
		Id             int64  `json:"id"`
		UserId         int64  `json:"-"`
		Url            string `json:"url"`
		ConnectionTime string `json:"connection_time"`
		PingTime       string `json:"ping_time"`
		User           User   `json:"-"`
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
		Url               string         `json:"url"`
		CountPing         int            `json:"count_ping"`
		CorrectCount      int            `json:"correct_count"`
		CancelCount       int            `json:"cancel_count"`
		MaxConnectionTime float64        `json:"max_connection_time"`
		MinConnectionTime float64        `json:"min_connection_time"`
		AvgConnectionTime float64        `json:"avg_connection_time"`
		Errors            []ErrorMessage `json:"errors,omitempty"`
	}

	ErrorMessage struct {
		Text  string `json:"text"`
		Count int    `json:"count"`
	}

	StatisticResultList []Statistic
)
