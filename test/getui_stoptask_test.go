package getui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"jawave.com/we-pm/message/getui"
)

// Test_StopTask 终止群发task
func Test_StopTask(t *testing.T) {
	init := getui.InitParams{
		AppID:         "你的appID",
		AppSecret:     "你的AppSecret",
		AppKey:        "你的appKey",
		MasterSecret:  "你的MasterSecret",
		AuthHeartbeat: 20, // 刷新时长，单位：小时
	}

	client, err := getui.GetClient(init)
	assert.Nil(t, err)

	rsp, err := client.StopTask("你的任务id")
	assert.Nil(t, err)
	assert.NotNil(t, rsp)
}
