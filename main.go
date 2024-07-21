package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

var urls = []string{
	"http://whatismyip.akamai.com/",
	"http://checkip.amazonaws.com",
	"https://checkip.amazonaws.com",
}

type DDNSRecord struct {
	Name         string
	TTL          int
	HostedZoneID string
	IP           string
}

func main() {
	log.Printf("start route53ddns")
	var timer = time.Now()
	var name string
	var ttl = 60
	var hostedZoneID string
	var resultText string
	var currentIP string
	var err error

	flag.StringVar(&name, "name", name, "Record hostname to update 'home.example.com'")
	flag.IntVar(&ttl, "ttl", ttl, "Record TTL in seconds")
	flag.StringVar(&hostedZoneID, "hostedZoneID", hostedZoneID, "Route53 Hosted Zone ID to update")
	flag.Parse()

	if strings.TrimSpace(name) == "" {
		log.Fatalf("name:%s not valid", name)
	}

	if strings.TrimSpace(hostedZoneID) == "" {
		log.Fatalf("hostedZoneID:%s not valid", hostedZoneID)
	}

	if ttl == 0 {
		log.Fatalf("ttl can not be 0")
	}

	for x, url := range urls {
		if currentIP, err = getCurrentIP(url); err != nil {
			log.Printf("getCurrentIP() error:%v", err)
			if x == len(url) {
				log.Fatalf("All IP check urls failed")
			}
		} else {
			break
		}
	}

	log.Printf("currentIP:%s", currentIP)

	if resultText, err = updateDNS(DDNSRecord{
		Name:         name,
		HostedZoneID: hostedZoneID,
		TTL:          ttl,
		IP:           currentIP,
	}); err != nil {
		log.Fatalf("updateDNS() error:%v", err)
	}

	log.Printf("resultText:%s", resultText)
	log.Printf("end route53ddns runtime:%s", time.Since(timer))
}

func updateDNS(record DDNSRecord) (resultText string, err error) {

	sess := session.Must(session.NewSession(&aws.Config{
		MaxRetries: aws.Int(3),
	}))

	svc := route53.New(sess)

	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("UPSERT"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(record.Name),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(record.IP),
							},
						},
						TTL:  aws.Int64(int64(record.TTL)),
						Type: aws.String("A"),
					},
				},
			},
			Comment: aws.String(fmt.Sprintf("DDNS Update for %s", record.Name)),
		},
		HostedZoneId: aws.String(record.HostedZoneID),
	}

	var result *route53.ChangeResourceRecordSetsOutput
	result, err = svc.ChangeResourceRecordSets(input)

	resultText = result.String()

	return resultText, err
}

func getCurrentIP(url string) (ip string, err error) {
	var resp *http.Response
	var body []byte

	var client = http.Client{
		Timeout: time.Second * 5,
	}

	if resp, err = client.Get(url); err != nil {
		err = fmt.Errorf("http.Get() url:%s error:%v", url, err)
		return ip,
			err
	}

	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		err = fmt.Errorf("ioutil.ReadAll() resp.Body error:%v", err)
		return ip, err
	}
	defer resp.Body.Close()

	ip = strings.TrimSpace(string(body))

	netIP := net.ParseIP(ip)
	if netIP == nil {
		err = fmt.Errorf("invalid IP:%s", ip)
		return ip, err
	}

	return ip, err
}
