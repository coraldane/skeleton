package proto

type LoginAuth struct {
	UserId int64  `json:"userId"`
	Token  string `json:"token"`
}
