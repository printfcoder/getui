package getui

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_GetAuth(t *testing.T) {

	ts := fmt.Sprintf("%d", int64(time.Now().UnixNano()/1000000))
	appID := "你的appID"
	appKey := "你的appKey"
	masterSecret := "你的MasterSecret"
	sign := sha256.Sum256([]byte(appKey + ts + masterSecret))
	signStr := fmt.Sprintf("%x", sign)
	t.Log(signStr)
	body := struct {
		AppKey    string `json:"appkey"`
		Timestamp string `json:"timestamp"`
		Sign      string `json:"sign"`
	}{AppKey: appKey, Timestamp: ts, Sign: signStr}

	data, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", "https://restapi.getui.com/v1/"+appID+"/auth_sign", ioutil.NopCloser(bytes.NewReader(data)))
	assert.Nil(t, err)
	req.Header.Add("Content-Type", "application/json")

	rsp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer rsp.Body.Close()

	rspBody, err := ioutil.ReadAll(rsp.Body)
	assert.Nil(t, err)

	ret := &struct {
		Result    string `json:"result"`
		AuthToken string `json:"auth_token"`
	}{}
	err = json.Unmarshal(rspBody, ret)
	assert.Nil(t, err)
	t.Log(ret.Result)
	t.Log(ret.AuthToken)
}
