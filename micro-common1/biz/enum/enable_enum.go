package enum

type EnableEnum uint8

const (
	Enable EnableEnum = iota + 1
	Disable
)

func (e EnableEnum) Ok() bool {
	return e == Enable
}
