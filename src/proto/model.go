package proto

type Entry struct {
	Term     int64
	ClientId int
	MsgId    int64
	Cmd      string
}
