package model

const (
	ProcessBefore ProcessType = "before"
	ProcessAfter  ProcessType = "after"
)

type (
	ProcessType string

	// CommandEvent структура представляет собой данные о команде которая была выполнена
	CommandEvent struct {
		Command string
		Process ProcessType
	}

	Emit struct {
	}
)
