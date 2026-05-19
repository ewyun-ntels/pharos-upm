package common

type CheckType string
type Type string

const (
	TypeQuery        Type = "query"
	TypeEventStatus  Type = "event-status"
	TypeEventHistory Type = "event-history"
)
