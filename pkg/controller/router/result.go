package router

const (
	SUCCESS int32 = 0
	ERROR   int32 = 1
)

type Response struct {
	Code int32       `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}
