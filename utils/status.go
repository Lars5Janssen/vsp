package utils

type ActiveStatus string

const (
	Active       ActiveStatus = "active"
	Left         ActiveStatus = "left"
	Disconnected ActiveStatus = "disconnected"
)

type MessagesStatus string

const (
	ACTIVE  MessagesStatus = "active"
	DELETED MessagesStatus = "deleted"
)
