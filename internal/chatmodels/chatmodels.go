package chatmodels

const (
	MsgTypeHello = "hello"
	MsgTypeChat  = "chat"
	MsgTypeJoin  = "join"
)

type Payload struct {
	MsgType string
	Nick    *string
	Msg     *string
}
