// 基于Go的阿里云DDNS工具
// 作者：青衿
// 主页：https://mengqinghe.com
// ip接口来源：https://cloud.tencent.com/developer/article/1152362
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"github.com/spf13/viper"
)

var client *alidns.Client

func init() {
	// 欢迎信息
	fmt.Println(`
	基于GO的阿里云DDNS工具
	作者：青衿
	主页：https://mengqinghe.com
	`)
	// 指定配置文件
	viper.SetConfigName("aliyunddns")
	viper.SetConfigType("json")
	viper.AddConfigPath("$HOME/aliyunddns/")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("启动失败: %s\n", err)
		os.Exit(1)
	}
}

func main() {
	domain := viper.Get("domain").(string)
	accessKeyID := viper.Get("accessKeyId").(string)
	accessSecret := viper.Get("accessSecret").(string)
	duration := viper.Get("duration").(float64)
	var publicIP string
	var Records []alidns.Record
	client, _ = alidns.NewClientWithAccessKey("cn-hangzhou", accessKeyID, accessSecret)
	timeTickChan := time.Tick(time.Hour * time.Duration(duration))
	for {
		publicIP = getPublicIP()
		Records = getSubDomainRecords(domain)
		fmt.Printf("当前公网IP地址是: %v\n", publicIP)
		if len(Records) == 0 {
			fmt.Println("当前域名未配置解析")
			return
		}
		if publicIP != Records[0].Value {
			fmt.Printf("当前解析IP地址为%s\n", Records[0].Value)
			fmt.Println("当前解析IP地址需要更新")
			upgradeDomainRecord(Records[0].RecordId, Records[0].RR, publicIP)
			fmt.Printf("当前解析IP地址已更新为: %s", publicIP)
		} else {
			fmt.Println("当前解析无需更新")
		}
		<-timeTickChan
	}

}

// 获取公网ip
func getPublicIP() string {
	res, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	return string(body)
}

// 获取域名解析记录, 暂只支持子域名
func getSubDomainRecords(Domain string) []alidns.Record {
	req := alidns.CreateDescribeSubDomainRecordsRequest()
	req.Scheme = "https"
	req.SubDomain = Domain
	res, err := client.DescribeSubDomainRecords(req)
	if err != nil {
		fmt.Print(err.Error())
	}
	return res.DomainRecords.Record
}

// 更新域名解析记录
func upgradeDomainRecord(RecordID string, RR string, IP string) {
	req := alidns.CreateUpdateDomainRecordRequest()
	req.Scheme = "https"
	req.RecordId = RecordID
	req.RR = RR
	req.Type = "A"
	req.Value = IP
	_, err := client.UpdateDomainRecord(req)
	if err != nil {
		fmt.Print(err.Error())
	}
}
