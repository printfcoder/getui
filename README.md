# getui
getui sdk for golang
目前最新版本 1.0.0

安装

     go get github.com/printfcoder/getui

鉴于GeTui 官网没有提供Golang的sdk包，这里简单封装一下常用的推送功能。封装采用单例模式


目前功能有：鉴权，单发，app推，终止群推，查询用户状态

未来{2017年4月25日后}
将持续开发：tolist推，批量单推，别名功能，标签管理，黑名单用户管理，其它推送结果与用户查询功能

重要说明

除初始化外，Appkey的设置将无效。
如何使用：
见tests目录下的各用例。


特别说明

对于所有推送请求，均采用post方式，个推会返回taskid，我会把它封装在rspBody中，供后续调用
