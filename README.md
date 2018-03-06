AliyunDDNS
===========================
A simple DDNS client written in Golang for Aliyun

Usage
-----------
1. Obtain `AccessKeyId ` and `AccessKeySecret ` from the [Aliyun AK Console][1]. For safety reasons, I recommend you use [Aliyun RAM console][2] to generate these parameters, but do not forget to grant the `AliyunDNSFullAccess` policy.
1. Edit `config.json` file according to your own configuration.
    1. `AccessKeyId` and `AccessKeySecret` are obtained in the first step.
    1. `Domain ` and `SubDomain` combined is the final domain name to be updated.
    1. `TTL` means the DNS record [TTL][3] value.
    1. `Duration` means the interval of each update. Also, When an update fails, it will retry after 60 seconds.
1. Run `go build aliyunddns.go` or download the executable for your platform from the [Releases Page][5].
1. Run the executable generated from the previous step.

License
-----------
The MIT License

****

阿里云 DDNS 客户端
===========================
由 Go 语言编写的简易阿里云 DDNS 客户端

使用方法
-----------
1. 从 [阿里云 AK 控制台][1] 取得 `AccessKeyId` 和 `AccessKeySecret`。为了安全起见，推荐使用 [阿里云 RAM 控制台][2] 生成该参数，但要注意赋予 `AliyunDNSFullAccess` 权限。
1. 根据自己的实际配置编辑 `config.json` 文件。
    1. 依次填入从第一步中取得的 `AccessKeyId ` 和 `AccessKeySecret `。
    1. `Domain ` 和 `SubDomain` 分别为域名和子域名，组成最终将要更新的域名。
    1. `TTL` 为 DNS 记录的 [TTL][4] 值。
    1. `Duration` 为每次更新 DNS 记录的间隔。注意，如果更新失败，将在60秒后重试。
1. 运行 `go build aliyunddns.go` 或从 [发布页面][5] 下载适用于你的平台的可执行文件。
1. 运行从上一步编译生成的可执行文件。

[1]: https://ak-console.aliyun.com/ "Aliyun AK Console"
[2]: https://ram.console.aliyun.com/#/user/list?guide "RAM console"
[3]: https://en.wikipedia.org/wiki/Time_to_live "Time-to-live"
[4]: https://zh.wikipedia.org/wiki/存活时间 "Time-to-live"
[5]: https://github.com/JerryLocke/AliyunDDNS/releases "Releases Page"