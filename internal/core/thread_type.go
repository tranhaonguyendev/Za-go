package core

type ThreadType int

const (
	USER ThreadType = iota
	GROUP
)

func (t ThreadType) String() string {
	switch t {
	case USER:
		return "USER"
	case GROUP:
		return "GROUP"
	default:
		return "UNKNOWN"
	}
}
