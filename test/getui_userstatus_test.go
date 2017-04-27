package getui

import (
	"testing"

	"github.com/printfcoder/getui"
	"github.com/stretchr/testify/assert"
)

// Test_UserStatus 查询用户状态
func Test_UserStatus(t *testing.T) {
	init := getui.InitParams{
		AppID:         "你的appID",
		AppSecret:     "你的AppSecret",
		AppKey:        "你的appKey",
		MasterSecret:  "你的MasterSecret",
		AuthHeartbeat: 20, // 刷新时长，单位：小时
	}

	client, err := getui.Init(init)
	assert.Nil(t, err)

	rsp, err := client.UserStatus("用户CID")
	assert.Nil(t, err)
	assert.NotNil(t, rsp)
}
