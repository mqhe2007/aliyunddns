// 基于Go的阿里云DDNS工具
// 作者：青衿
// 主页：https://mengqinghe.com
// ip接口来源：https://cloud.tencent.com/developer/article/1152362
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"github.com/jinzhu/configor"
	"github.com/robfig/cron"
)

// Config 配置文件结构体
type Config struct {
	RegionID     string
	Domain       string
	AccessKeyID  string
	AccessSecret string
	Cron         string
}

var (
	configFlag        string     // 配置文件地址参数
	logFlag           string     // 日志文件地址参数
	loopCount         int    = 1 // 更新循环次数统计
	config            Config
	aliyunClient      *alidns.Client
	aliyunClientError error
)

func init() {
	// 欢迎信息
	fmt.Println(`
	基于GO的阿里云DDNS工具
	作者：青衿
	主页：https://mengqinghe.com
	`)
	// 命令行参数接收
	flag.StringVar(&configFlag, "c", "./aliyunddns-config.json", "指定配置文件")
	flag.StringVar(&logFlag, "l", "./aliyunddns-out.log", "指定日志文件")
	flag.Parse()
	// 日志文件设置
	logFile, openFileErr := os.OpenFile(logFlag, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if openFileErr != nil {
		fmt.Printf("打开日志文件错误：%v", openFileErr)
		os.Exit(0)
	}
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime)
}

func main() {
	// 读取配置文件
	readConfigErr := readConfig()
	if readConfigErr != nil {
		log.Printf("打开配置文件错误：%v", readConfigErr)
		return
	}
	fmt.Println("配置文件：", configFlag)
	fmt.Println("日志文件：", logFlag)
	domainName, RR := parseDomain(config.Domain)
	// 创建阿里云解析API客户端
	aliyunClient, aliyunClientError = alidns.NewClientWithAccessKey(config.RegionID, config.AccessKeyID, config.AccessSecret)
	if aliyunClientError != nil {
		log.Panic(aliyunClientError)
	}
	// 定时器开始前先干一次。
	doIt(domainName, RR)
	c := cron.New()
	c.AddFunc(config.Cron, func() {
		doIt(domainName, RR)
	})
	c.Start()
	select {}
}

// 读取配置文件
func readConfig() (err error) {
	err = configor.New(&configor.Config{ErrorOnUnmatchedKeys: true}).Load(&config, configFlag)
	log.Printf("配置文件：%#v", config)
	return
}

// 解析配置项中的域名，解析出主域名和RR值
func parseDomain(configDomain string) (domainName string, RR string) {
	domainLevel := strings.Count(configDomain, ".")
	const (
		_ = iota
		split1
		splitMore
	)
	if domainLevel == 1 {
		domainSlice := strings.SplitN(configDomain, ".", split1)
		domainName = domainSlice[0]
		RR = "@"
	} else {
		domainSlice := strings.SplitN(configDomain, ".", splitMore)
		domainName = domainSlice[1]
		RR = domainSlice[0]
	}
	return
}

// 获取当前环境公网IP
func getPublicIP() (publicIP string) {
	res, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	publicIP = string(body)
	log.Print("当前环境的公网IP：", publicIP)
	return
}

// 获取域名解析记录
func getDomainRecords(domainName string, RR string) []alidns.Record {
	req := alidns.CreateDescribeDomainRecordsRequest()
	req.DomainName = domainName
	req.KeyWord = RR
	req.Type = "A"
	req.SearchMode = "EXACT"
	res, err := aliyunClient.DescribeDomainRecords(req)
	if err != nil {
		log.Print(err.Error())
	}
	log.Printf("当前解析记录：%#v", res.DomainRecords.Record)
	return res.DomainRecords.Record
}

// 创建域名解析记录
func addDomainRecord(domainName string, RR string, value string) (recordID string) {
	req := alidns.CreateAddDomainRecordRequest()
	req.DomainName = domainName
	req.RR = RR
	req.Type = "A"
	req.Value = value
	res, err := aliyunClient.AddDomainRecord(req)
	if err != nil {
		log.Print(err.Error())
	}
	recordID = res.RecordId
	log.Printf("新建解析记录ID：%v", recordID)
	return
}

// 更新域名解析记录
func upgradeDomainRecord(recordID string, RR string, value string) {
	req := alidns.CreateUpdateDomainRecordRequest()
	req.RecordId = recordID
	req.RR = RR
	req.Type = "A"
	req.Value = value
	_, err := aliyunClient.UpdateDomainRecord(req)
	if err != nil {
		log.Print(err.Error())
	}
	log.Printf("当前域名解析记录值已更新为: %v", value)
}

// 干它！
func doIt(domainName, RR string) {
	log.Printf("第%v次更新开始-----------------------------", loopCount)
	// 获取当前环境的公网IP
	publicIP := getPublicIP()
	alidnsRecords := getDomainRecords(domainName, RR)
	if len(alidnsRecords) == 0 {
		log.Print("当前目标域名解析记录不存在，接下来会自动创建。")
		addDomainRecord(domainName, RR, publicIP)
	} else {
		targetRecord := alidnsRecords[0]
		if publicIP != targetRecord.Value {
			log.Printf("当前解析IP地址%v需要更新", targetRecord.Value)
			upgradeDomainRecord(targetRecord.RecordId, targetRecord.RR, publicIP)
		} else {
			log.Println("当前解析记录值无需更新")
		}
	}
	log.Printf("第%v次更新结束+++++++++++++++++++++++++++++", loopCount)
	loopCount++
}
