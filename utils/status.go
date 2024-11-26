package utils

type ComponentStatus int

const (
	OK                 ComponentStatus = 200
	Unauthorized       ComponentStatus = 401
	Forbidden          ComponentStatus = 403
	NotFound           ComponentStatus = 404
	Conflict           ComponentStatus = 409
	PreconditionFailed ComponentStatus = 412
)

type ActiveStatus string

const (
	Active       ActiveStatus = "active"
	Left         ActiveStatus = "left"
	Disconnected ActiveStatus = "disconnected"
)
