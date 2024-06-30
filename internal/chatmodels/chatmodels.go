package chatmodels

const (
	MsgTypeHello = "hello"
	MsgTypeChat  = "chat"
	MsgTypeJoin  = "join"
	MsgTypeAnn   = "announcement"
	MsgTypeDM    = "dm"
)

type Payload struct {
	MsgType string
	Nick    *string
	Msg     *string
}
