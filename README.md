# aliyunddns 基于 Go 的阿里云 DDNS 工具

> 任何问题，请在 issues 反馈。

## 概述

使用 aliyunddns 工具，可以定时更新（[cron](https://zh.wikipedia.org/wiki/Cron)）您在阿里云的域名解析服务。

## 命令行参数

### -h

查看帮助

### -c

默认值：`./aliyunddns-config.json`

指定配置文件。

### -l

默认值：`./aliyunddns-out.log`

指定日志输出文件。

## 教程

### 1. 下载

在 release 页面下载对应平台的可执行程序。

### 2. 创建配置文件

创建配置文件，一个满足`Config`结构体的`.json`文件，如果不通过命令行参数指定配置文件，请在程序目录创建名为`aliyunddns-config.json`的文件。

```go
type Config struct {
  // 区域ID，因本工具不涉及ECS等实体产品操作，此字段对运行无影响，统一设置为"cn-hangzhou"即可
  // 详细说明参见：https://help.aliyun.com/document_detail/40654.html
  RegionID     string
  // 支持子域名，当值为主域名时，默认添加主机记录"@"
  Domain       string
  // 如何获取AccessKeyID和AccessSecret请参见：https://ram.console.aliyun.com/users/new
  AccessKeyID  string
  AccessSecret string
  // CRON表达式格式参见：https://pkg.go.dev/github.com/robfig/cron
  Cron         string
}
```

### 3. 后台运行
