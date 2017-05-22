package getui

// SaveListBody 保存消息共同体
// list推时有需要先把消息共同体保存到个推，再发送推送到客户端的请求
type SaveListBody struct {
	Message      saveListBodymessage `json:"message"`
	Notification Notification        `json:"notification"`
}

type saveListBodymessage struct {
	AppKey            string `json:"appkey"`
	IsOffLine         bool   `json:"is_offline"`
	OfflineExpireTime int64  `json:"offline_expire_time"`
	MsgType           string `json:"msgtype"`
}
