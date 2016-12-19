package main

import (
	_ "errors"
	"flag"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/ghodss/yaml"
)

type S3Server struct {
	Region   string `yaml:"region"`
	ID       string
	Secret   string
	Endpoint string
	Signver  string
	SizeList string `yaml:"sizelist"`
	Count    int
}

func main() {

	bucket := flag.String("bucket", "", "the name of an existing bucket")
	num := flag.Int("num", 1, "the number of objects to be created")
	size := flag.Int("size", 1, "size of objects (KB)")
	randomsize := flag.Bool("randomsize", false, "Random object size from 1K to 10M")
	svrname := flag.String("server", "default", "S3 Server (from config.yml) to contact with")

	flag.Parse()

	//Load server configuration
	svr, err := loadCfg(*svrname)
	if err != nil {
		panic(err)
	}
	//fmt.Println(*svr)
	fmt.Println(*bucket, *num, *size, *randomsize)

	//create s3 service
	s3svc := s3.New(session.New(), &aws.Config{
		Endpoint:         aws.String(svr.Endpoint),
		Credentials:      credentials.NewStaticCredentials(svr.ID, svr.Secret, ""),
		Region:           &svr.Region,
		S3ForcePathStyle: aws.Bool(svr.Endpoint != ""),
	})

	sl := strings.ToUpper(svr.SizeList)

	sa := make([]int64, len(strings.Split(sl, ",")))
	for idx, s := range strings.Split(sl, ",") {
		l := len(s)
		sz, _ := strconv.Atoi(s[:l-1])
		if s[l-1] == 'K' {
			sz = sz * 1024
		} else if s[l-1] == 'M' {
			sz = sz * 1024 * 1024
		} else if s[l-1] == 'G' {
			sz = sz * 1024 * 1024 * 1024
		} else if s[l-1] == 'T' {
			sz = sz * 1024 * 1024 * 1024 * 1024
		}

		sa[idx] = int64(sz)
		idx++
	}

	fmt.Println(sa)

	ps, _ := NewPerfStats(sa)

	var wg sync.WaitGroup
	wg.Add(len(sa))
	for _, sz := range sa {
		fmt.Println("Start uploading...", sz, "*", svr.Count)
		uploadRandomObj(&wg, ps, s3svc, "test-zhaoy1", sz, svr.Count)
	}

	wg.Wait()

	fmt.Println("Done uploading")
	ps.PrintStats()

}

//Load configuration
func loadCfg(server string) (*S3Server, error) {
	cfg, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		panic("Load configuration failed")
	}

	var servers map[string]S3Server
	err = yaml.Unmarshal(cfg, &servers)
	if err != nil {
		panic("Parse configuration failed")
	}

	srv, prs := servers[server]
	if !prs {
		panic("No definition for specified server")
	}

	return &srv, err
}

//upload a bunch of random object with the specified size
func uploadRandomObj(wg *sync.WaitGroup, ps *PerfStats, svc *s3.S3, bucket string, size int64, count int) error {

	defer wg.Done()

	//TODO: create bucket if it doesn't exist

	for i := 0; i < int(count); i++ {
		obj, err := NewRandomObject("", "Test", size)
		if err != nil {
			panic("Failed to new random object!")
		}

		params := &s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(obj.Key),
			Body:   obj,
		}

		t := time.Now()

		_, err2 := svc.PutObject(params)
		if err2 != nil {
			panic("Failed to upload object" + err2.Error())
		}

		ps.Add1Duration(size, (time.Now().Sub(t)).Nanoseconds()/1000000)
	}

	return nil
}
