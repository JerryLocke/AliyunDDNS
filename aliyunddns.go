package main

import (
	"fmt"
	"io/ioutil"
	"encoding/json"
	"time"
	"net/http"
	"sort"
	"bytes"
	"net/url"
	"strings"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
)

const (
	ProgramName    = "Aliyun DDNS Tools"
	ProgramAuthor  = "@JerryLocke"
	ProgramVersion = "v1.0.0"
	ProgramInfo    = ProgramName + "\n" + ProgramAuthor + "\n" + ProgramVersion + "\n"

	AliDnsApi = "http://alidns.aliyuncs.com/?"
)

var lastIp string

type Config struct {
	AccessKeyId     string
	AccessKeySecret string
	Domain          string
	SubDomain       string
	TTL             int
	Duration        int
}

type IpInfo struct {
	Ip string `json:"ip"`
}

type DomainRecord struct {
	RR         string
	Value      string
	RecordId   string
	Type       string
	DomainName string
}

type DomainList struct {
	TotalCount int
	PageNumber int
	PageSize   int
	DomainRecords struct {
		Record []DomainRecord
	}
}

type UpdateResult struct {
	Code    string
	Message string
}

func main() {
	fmt.Println(ProgramInfo)
	CheckRecord()
	select {}
}

func CheckRecord() {
	PrintLog("Checking IP...")
	config, err := ReadConfig()
	if err != nil {
		PrintLog("Read config failed: %v", err)
		NextCheck(nil)
		return
	}
	if config.AccessKeyId == "" || config.AccessKeySecret == "" || config.Domain == "" || config.SubDomain == "" || config.TTL == 0 || config.Duration == 0 {
		PrintLog("Invalid config!")
		NextCheck(nil)
		return
	}
	ip, err := GetMyIp()
	if err != nil {
		PrintLog("Get current IP failed: %v", err)
		NextCheck(nil)
		return
	}
	PrintLog("Current IP：%v，last IP：%v", ip, lastIp)
	if ip != lastIp {
		PrintLog("Looking for old DNS record...")
		record := FindRecord(config)
		if record.Value == ip {
			PrintLog("The record does not need to be updated")
		} else {
			if record == nil {
				PrintLog("No old record found, will add new DNS record")
			}
			PrintLog("Updating record...")
			result, err := UpdateRecord(record.RecordId, ip, config)
			if err != nil {
				PrintLog("Update failed：%v", err)
				NextCheck(nil)
				return
			}
			if result.Code == "" {
				PrintLog("Update succeeded")
			} else {
				PrintLog("Update failed：code=%v, message", result.Code, result.Message)
				NextCheck(nil)
				return
			}
		}
		lastIp = ip
	} else {
		PrintLog("Already up-to-date")
	}
	NextCheck(config)
}

func ReadConfig() (*Config, error) {
	bs, err := ioutil.ReadFile("config.json")
	if err != nil {
		return nil, err
	}
	config := &Config{}
	err = json.Unmarshal(bs, &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func NextCheck(config *Config) {
	duration := time.Duration(60)
	if config != nil {
		duration = time.Duration(config.Duration)
	}
	PrintLog("Program Will be executed again after %d seconds\n", duration)
	time.AfterFunc(duration*time.Second, CheckRecord)
}

func GetMyIp() (string, error) {
	ipInfo := &IpInfo{}
	err := GetJson("http://ipinfo.io/json", ipInfo)
	if err != nil {
		return "", err
	}
	return ipInfo.Ip, nil
}

func FindRecord(config *Config) *DomainRecord {
	for i := 1; ; {
		list, err := GetDomainList(config, i)
		if err != nil {
			break
		}
		if list.TotalCount == 0 || len(list.DomainRecords.Record) == 0 {
			break
		}
		for _, domain := range list.DomainRecords.Record {
			if domain.DomainName == config.Domain && domain.RR == config.SubDomain && domain.Type == "A" {
				return &domain
			}
		}
	}
	return nil
}

func GetDomainList(config *Config, page int) (*DomainList, error) {
	param := make(map[string]string)
	param["Action"] = "DescribeDomainRecords"
	param["DomainName"] = config.Domain
	param["PageNumber"] = fmt.Sprintf("%d", page)
	param["PageSize"] = "500"
	queryString := BuildQueryString(config, param)

	list := &DomainList{}
	err := GetJson(AliDnsApi+queryString, list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func UpdateRecord(recordId string, ip string, config *Config) (*UpdateResult, error) {
	param := make(map[string]string)
	if recordId == "" {
		param["Action"] = "AddDomainRecord"
		param["DomainName"] = config.Domain
	} else {
		param["Action"] = "UpdateDomainRecord"
		param["RecordId"] = recordId
	}
	param["RR"] = config.SubDomain
	param["Type"] = "A"
	param["Value"] = ip
	param["TTL"] = fmt.Sprintf("%d", config.TTL)
	queryString := BuildQueryString(config, param)
	result := &UpdateResult{}
	err := GetJson(AliDnsApi+queryString, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func GetJson(url string, object interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if object != nil {
		err = json.Unmarshal(body, object)
		if err != nil {
			return err
		}
	}
	return nil
}

func BuildQueryString(config *Config, params map[string]string) string {
	params["Format"] = "json"
	params["Version"] = "2015-01-09"
	params["AccessKeyId"] = config.AccessKeyId
	params["SignatureMethod"] = "HMAC-SHA1"

	now := time.Now()
	utc := now.UTC()
	year, mon, day := utc.Date()
	hour, min, sec := utc.Clock()
	params["Timestamp"] = fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02dZ",
		year, mon, day, hour, min, sec)

	params["SignatureVersion"] = "1.0"
	params["SignatureNonce"] = fmt.Sprintf("%d", utc.UnixNano())

	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var stringToSign bytes.Buffer
	var canonicalizedQueryString bytes.Buffer
	const SEPARATOR = "&"
	stringToSign.WriteString("GET")
	stringToSign.WriteString(SEPARATOR)
	stringToSign.WriteString(PercentEncode("/"))
	stringToSign.WriteString(SEPARATOR)
	for _, k := range keys {
		if k != "" {
			canonicalizedQueryString.WriteString(SEPARATOR)
			canonicalizedQueryString.WriteString(PercentEncode(k))
			canonicalizedQueryString.WriteString("=")
			canonicalizedQueryString.WriteString(PercentEncode(params[k]))
		}
	}
	stringToSign.WriteString(PercentEncode(canonicalizedQueryString.String()[1:]))
	params["Signature"] = HmacSHA1([]byte(stringToSign.String()), config.AccessKeySecret+SEPARATOR)

	var queryString bytes.Buffer
	for k := range params {
		queryString.WriteString(SEPARATOR)
		queryString.WriteString(k)
		queryString.WriteString("=")
		queryString.WriteString(PercentEncode(params[k]))
	}
	return queryString.String()[1:]
}

func PercentEncode(value string) string {
	value = url.QueryEscape(value)
	value = strings.Replace(value, "=", "%3D", -1)
	value = strings.Replace(value, "+", "%20", -1)
	value = strings.Replace(value, "*", "%2A", -1)
	value = strings.Replace(value, "%7E", "~", -1)
	return value
}

func HmacSHA1(message []byte, key string) string {
	signature := hmac.New(sha1.New, []byte(key))
	signature.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(signature.Sum(nil))
}

func PrintLog(format string, a ...interface{}) {
	now := time.Now()
	year, mon, day := now.Date()
	hour, min, sec := now.Clock()
	message := format
	if a != nil {
		message = fmt.Sprintf(format, a...)
	}
	fmt.Printf("[%04d-%02d-%02d %02d:%02d:%02d]%v\n",
		year, mon, day, hour, min, sec, message)
}
