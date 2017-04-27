package getui

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// Message 请求消息配置 Message
type Message struct {
	AppKey    string `json:"appkey"`
	IsOffline bool   `json:"is_offline"`
	MsgType   string `json:"msgtype"`
}

// Notification 请求消息配置 Notification
// 资料 http://docs.getui.com/server/rest/template/
type Notification struct {
	Style struct {
		Type  int    `json:"type"`
		Text  string `json:"text"`
		Title string `json:"title"`
	} `json:"style"`
	TransmissionType    bool   `json:"transmission_type"`
	TransmissionContent string `json:"transmission_content"`
	// 带duration的有bug，貌似不会显示
	// DurationBegin       string `json:"duration_begin,omitempty"`
	// DurationEnd         string `json:"duration_end,omitempty"`
}

// SingleReqBody 个推请求body 单推
// 参考资料 http://docs.getui.com/server/rest/push/#3
type SingleReqBody struct {
	Message      Message      `json:"message"`
	Notification Notification `json:"notification"`
	CID          string       `json:"cid,omitempty"`
	Alias        string       `json:"alias,omitempty"`
	RequestID    string       `json:"requestid"`
}

// AppReqBody 个推请求body toapp
// 参考资料 http://docs.getui.com/server/rest/push/#5-toapp
type AppReqBody struct {
	Message      Message               `json:"message"`
	Notification Notification          `json:"notification"`
	Condition    []AppReqBodyCondition `json:"condition"`
	RequestID    string                `json:"requestid"`
}

// AppReqBodyCondition toapp 过滤条件
// 参考资料 http://docs.getui.com/server/rest/push/#5-toapp
type AppReqBodyCondition struct {
	Key     string   `json:"key"`
	Values  []string `json:"values"`
	OptType string   `json:"opt_type"`
}

// RspBody 个推Rsp body
// 个推请求返回的结构
// status : successed_offline 离线下发
//          successed_online 在线下发
//          successed_ignore 非活跃用户不下发
type RspBody struct {
	Result    string `json:"result"`
	TaskID    string `json:"taskid"`
	Desc      string `json:"desc"`
	Status    string `json:"status"`
	RequestID string `json:"requestID,omitempty"`
}

// UserStatus 用户状态 rsp body
type UserStatus struct {
	Result        string `json:"result"`
	CID           string `json:"cid"`
	Status        string `json:"status"`
	LastLoginUnix int64  `json:"lastlogin"`
	LastLogin     time.Time
}

// Client 客户端接口
type Client interface {
	PushToSingle(SingleReqBody) (*RspBody, error)
	PushToApp(AppReqBody) (*RspBody, error)
	StopTask(string) (*RspBody, error)
	UserStatus(string) (*UserStatus, error)
	CloseAuth() (*RspBody, error)
}

// InitParams 初始化参数
type InitParams struct {
	AppID        string
	AppSecret    string
	AppKey       string
	MasterSecret string
	// AuthHeartbeat Auth刷新时间 单位小时 默认20小时
	AuthHeartbeat time.Duration
}

type client struct {
	InitParams
	lastUpdateTokenTime time.Time
	authToken           string
}

var single *client

// Init 客户端-单例
func Init(parms InitParams) (c Client, err error) {
	if single == nil {
		single = new(client)
		single.AppID = parms.AppID
		single.AppSecret = parms.AppSecret
		single.AppKey = parms.AppKey
		single.MasterSecret = parms.MasterSecret
		single.AuthHeartbeat = parms.AuthHeartbeat

		err = single.init()
		if err != nil {
			return nil, fmt.Errorf("[GetClient] 初始化失败，err: %s", err)
		}

	}
	return single, nil
}

func (c *client) init() (err error) {

	// 申请token
	err = c.refreshAuth()
	if err != nil {
		return err
	}

	// 定时刷新token
	go func() {
		if c.AuthHeartbeat == 0 {
			c.AuthHeartbeat = 20
		}

		timer := time.NewTicker(c.AuthHeartbeat * time.Hour)
		for t := range timer.C {
			c.lastUpdateTokenTime = t
			c.refreshAuth()
		}

		select {}
	}()

	return nil
}

// refreshAuth 刷新认证，默认20小时一次
func (c *client) refreshAuth() error {

	// 有token则先清除掉
	if len(c.authToken) > 0 {
		_, err := c.CloseAuth()
		if err != nil {
			return fmt.Errorf("[refreshAuth] 关闭json，失败,err:%s", err)
		}
	}

	// 请求authToken
	// 参数构造
	ts := fmt.Sprintf("%d", int64(time.Now().UnixNano()/1000000))
	sign := sha256.Sum256([]byte(c.AppKey + ts + c.MasterSecret))
	signStr := fmt.Sprintf("%x", sign)
	body := struct {
		AppKey    string `json:"appkey"`
		Timestamp string `json:"timestamp"`
		Sign      string `json:"sign"`
	}{AppKey: c.AppKey, Timestamp: ts, Sign: signStr}
	data, _ := json.Marshal(body)

	// 创建请求
	req, err := http.NewRequest("POST", "https://restapi.getui.com/v1/"+c.AppID+"/auth_sign", ioutil.NopCloser(bytes.NewReader(data)))
	if err != nil {
		return fmt.Errorf("[refreshAuth] 创建auth请求失败, err: %s", err)
	}
	req.Header.Add("Content-Type", "application/json")

	// 发送请求
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("[refreshAuth] 发送auth请求失败, err: %s", err)
	}
	defer rsp.Body.Close()

	// 解析-body
	rspBody, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return fmt.Errorf("[refreshAuth] 发送auth请求返回的body无法解析, err: %s", err)
	}

	// 解析-JSON
	ret := &struct {
		Result    string `json:"result"`
		AuthToken string `json:"auth_token"`
	}{}
	err = json.Unmarshal(rspBody, ret)
	if err != nil {
		return fmt.Errorf("[refreshAuth] 发送auth请求返回的JSON无法解析, err: %s", err)
	}

	// 将token放到实例中
	c.authToken = ret.AuthToken

	return nil
}

// CloseAuth 清空Auth
func (c *client) CloseAuth() (ret *RspBody, err error) {
	req, err := http.NewRequest("POST", "https://restapi.getui.com/v1/"+c.AppID+"/auth_close", nil)
	if err != nil {
		return nil, fmt.Errorf("[CloseAuth] 创建 清空auth 请求失败, err: %s", err)
	}

	req.Header["authtoken"] = []string{c.authToken}
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[CloseAuth] 发送 清空auth 请求失败, err: %s", err)
	}
	defer rsp.Body.Close()

	rspBody, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, fmt.Errorf("[CloseAuth] 清空auth 请求返回的body无法解析, err: %s", err)
	}

	ret = &RspBody{}
	err = json.Unmarshal(rspBody, ret)
	if err != nil {
		return nil, fmt.Errorf("[CloseAuth] 清空auth 请求返回的JSON无法解析, err: %s", err)
	}

	if ret.Result != "ok" {
		return nil, fmt.Errorf("[CloseAuth] 清空auth 失败, desc: %s", ret.Desc)
	}

	return
}

// PushToSingle 发送单条信息
// 参考资料 http://docs.getui.com/server/rest/push/#3
func (c *client) PushToSingle(body SingleReqBody) (ret *RspBody, err error) {

	if len(body.CID) == 0 && len(body.Alias) == 0 {
		return nil, fmt.Errorf("[PushToSingle] 错误的目标设备, cid 与 alias 任选且必选一个")
	}

	body.Message.AppKey = c.AppKey
	if len(body.RequestID) == 0 {
		body.RequestID = strconv.FormatInt(time.Now().UnixNano(), 12)
	}

	// 构造请求
	data, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", "https://restapi.getui.com/v1/"+c.AppID+"/push_single", ioutil.NopCloser(bytes.NewReader(data)))
	if err != nil {
		return nil, fmt.Errorf("[PushToSingle] 创建 发送单条信息 请求失败, err: %s", err)
	}

	req.Header["Content-Type"] = []string{"application/json"}
	req.Header["authtoken"] = []string{c.authToken}

	// 发送请求
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[PushToSingle] 发送 单条信息 请求失败, err: %s", err)
	}
	defer rsp.Body.Close()

	// 解析-body
	rspBody, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, fmt.Errorf("[PushToSingle] 发送 单条信息请求 返回的body无法解析, err: %s", err)
	}

	// 解析-json
	ret = &RspBody{
		RequestID: body.RequestID,
	}
	err = json.Unmarshal(rspBody, ret)
	if err != nil {
		return nil, fmt.Errorf("[PushToSingle] 发送 单条信息 请求返回的JSON无法解析, err: %s", err)
	}

	if ret.Result != "ok" {
		return nil, fmt.Errorf("[PushToSingle] 发送 单条信息 请求不成功, ret: %v", ret)
	}

	return
}

// Push 向app推送
// 参考资料 http://docs.getui.com/server/rest/push/#5-toapp
func (c *client) PushToApp(body AppReqBody) (ret *RspBody, err error) {

	body.Message.AppKey = c.AppKey
	if len(body.RequestID) == 0 {
		body.RequestID = strconv.FormatInt(time.Now().UnixNano(), 12)
	}

	// 构造请求
	data, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", "https://restapi.getui.com/v1/"+c.AppID+"/push_app", ioutil.NopCloser(bytes.NewReader(data)))
	if err != nil {
		return nil, fmt.Errorf("[PushToSingle] 创建 向app推送信息 请求失败, err: %s", err)
	}

	req.Header["Content-Type"] = []string{"application/json"}
	req.Header["authtoken"] = []string{c.authToken}

	// 发送请求
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[PushToSingle] 发送 向app推送信 息请求失败, err: %s", err)
	}
	defer rsp.Body.Close()

	// 解析-body
	rspBody, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, fmt.Errorf("[PushToSingle] 发送 向app推送信息 请求返回的body无法解析, err: %s", err)
	}

	// 解析-json
	ret = &RspBody{
		RequestID: body.RequestID,
	}
	err = json.Unmarshal(rspBody, ret)
	if err != nil {
		return nil, fmt.Errorf("[PushToSingle] 发送 向app推送信息 请求返回的JSON无法解析, err: %s", err)
	}

	if ret.Result != "ok" {
		return nil, fmt.Errorf("[PushToSingle] 发送 向app推送信息 请求不成功, ret: %v ", ret)
	}

	return
}

// StopTask 终止群推任务
// 参考资料 http://docs.getui.com/server/rest/push/#6-stop
func (c *client) StopTask(taskID string) (ret *RspBody, err error) {

	req, err := http.NewRequest("DELETE", "https://restapi.getui.com/v1/"+c.AppID+"/stop_task/"+taskID, nil)
	if err != nil {
		return nil, fmt.Errorf("[StopTask] 创建 终止群推任务 信息请求失败, err: %s", err)
	}

	req.Header["Content-Type"] = []string{"application/json"}
	req.Header["authtoken"] = []string{c.authToken}

	// 发送请求
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[StopTask] 发送 终止群推任务 信息请求失败, err: %s", err)
	}
	defer rsp.Body.Close()

	// 解析-body
	rspBody, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, fmt.Errorf("[StopTask] 发送 终止群推任务 信息请求返回的body无法解析, err: %s", err)
	}

	// 解析-json
	ret = &RspBody{}
	err = json.Unmarshal(rspBody, ret)
	if err != nil {
		return nil, fmt.Errorf("[StopTask] 发送 终止群推任务 信息请求返回的JSON无法解析, err: %s", err)
	}

	if ret.Result != "ok" {
		return nil, fmt.Errorf("[StopTask] 发送 终止群推任务 信息请求不成功, ret: %v", ret)
	}

	return
}

// UserStatus 查看用户状态
// 参考资料 http://docs.getui.com/server/rest/push/#11_1
func (c *client) UserStatus(cid string) (ret *UserStatus, err error) {

	req, err := http.NewRequest("GET", "https://restapi.getui.com/v1/"+c.AppID+"/user_status/"+cid, nil)
	if err != nil {
		return nil, fmt.Errorf("[UserStatus] 创建 查看用户状态 请求失败, err: %s", err)
	}

	req.Header["Content-Type"] = []string{"application/json"}
	req.Header["authtoken"] = []string{c.authToken}

	// 发送请求
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[UserStatus] 发送 查看用户状态 请求失败, err: %s", err)
	}
	defer rsp.Body.Close()

	// 解析-body
	rspBody, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, fmt.Errorf("[UserStatus] 发送 查看用户状态 请求返回的body无法解析, err: %s", err)
	}

	// 解析-json
	ret = &UserStatus{}
	err = json.Unmarshal(rspBody, ret)
	if err != nil {
		return nil, fmt.Errorf("[UserStatus] 发送 查看用户状态 返回的JSON无法解析, err: %s", err)
	}

	ret.LastLogin = time.Unix(ret.LastLoginUnix/1000, 0)

	if ret.Result != "ok" {
		return nil, fmt.Errorf("[UserStatus] 发送 查看用户状态 请求不成功, ret: %v", ret)
	}

	return
}
