package getui

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test_CloseAuth 关闭token
func Test_CloseAuth(t *testing.T) {
	appID := "你的appID"
	authtoken := "你的authtoken"

	req, err := http.NewRequest("POST", "https://restapi.getui.com/v1/"+appID+"/auth_close", nil)
	assert.Nil(t, err)
	req.Header["authtoken"] = []string{authtoken}

	rsp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer rsp.Body.Close()

	rspBody, err := ioutil.ReadAll(rsp.Body)
	assert.Nil(t, err)

	ret := &struct {
		Result string `json:"result"`
	}{}
	err = json.Unmarshal(rspBody, ret)
	assert.Nil(t, err)
	t.Log(ret.Result)
}
