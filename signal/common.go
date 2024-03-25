package signal

type Result[U any] struct {
	Value U
	Err   error
}

type UserInfo struct {
	Name  string
	Token string
}

type Size struct {
	Width  int
	Height int
}

type HomeTabSelected bool

type Refetch string

type Connect struct {
	IsRoom bool
	Value  string
	Password string
}
