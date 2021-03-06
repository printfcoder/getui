package getui

import (
	"testing"

	"github.com/printfcoder/getui"
	"github.com/stretchr/testify/assert"
)

// Test_Single 单个用户
func Test_Single(t *testing.T) {
	init := getui.InitParams{
		AppID:         "你的appID",
		AppSecret:     "你的AppSecret",
		AppKey:        "你的appKey",
		MasterSecret:  "你的MasterSecret",
		AuthHeartbeat: 20, // 刷新时长，单位：小时
	}

	client, err := getui.Init(init)
	assert.Nil(t, err)

	reqBody := getui.SingleReqBody{}
	reqBody.Message.IsOffline = false
	reqBody.Message.MsgType = "notification"
	reqBody.Notification.Style.Text = "这是TextAps内2容"
	reqBody.Notification.Style.Type = 0
	reqBody.Notification.Style.Title = "这是titl2se"
	reqBody.Notification.TransmissionType = true
	reqBody.Notification.TransmissionContent = "透传内容"
	reqBody.CID = "你的CID"
	rsp, err := client.PushToSingle(reqBody)
	assert.Nil(t, err)
	assert.NotNil(t, rsp)
}
