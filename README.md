## Todo
* 解决原有gcron只能单机执行
* 脚本机状态上报
* alarm告警作为公共组件
* 将公共组件集中管理


## scron使用
### 第一步 本地go环境
``` go env
"GOPRIVATE": "github.com/olaola-chat/",
"GONOPROXY": "github.com/olaola-chat/",
"GONOSUMDB": "github.com/olaola-chat/",
```
### 第二步 拉取依赖
```mod
go get -u github.com/olaola-chat/slp-tools
go mod tidy // 最好是更新一下依赖关系
```
### 第三步 使用
```
import "github.com/olaola-chat/slp-tools/scron"
 
cron := scron.new()
if _, err := cron.AddSingleton("*/10 * * * * *", TestCronTab, "act.craving.run.TestCronTab"); err != nil {
	}
cron.start()
```
* 和使用原有gcron 注册cron方法一样,底层的变动对业务无感
