package main

import (
	_ "errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
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
	svrname := flag.String("server", "default", "S3 Server (from config.yml) to contact with")
	//_ := flag.Int("num", 1, "the number of objects to be created")
	//_ := flag.Bool("randomsize", false, "Random object size from 1K to 10M")
	flag.Parse()

	//Load server configuration
	svr, err := loadCfg(*svrname)
	if err != nil {
		panic(err)
	}

	//create s3 service
	cfg := &aws.Config{
		Endpoint:         aws.String(svr.Endpoint),
		Credentials:      credentials.NewStaticCredentials(svr.ID, svr.Secret, ""),
		Region:           aws.String(svr.Region),
		S3ForcePathStyle: aws.Bool(true),
		DisableSSL:       aws.Bool(false),
	}
	s3svc := s3.New(session.New(cfg))

	//Create bucket if doesn't exist
	if *bucket == "" {
		*bucket = "s3-perf-test"
	}
	err = CreateBucket(s3svc, bucket)
	if err != nil {
		panic("Fail to create bucket: " + err.Error())
	}

	//Get the list of the object sizes
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
		} else {
			sz, _ = strconv.Atoi(s)
		}

		sa[idx] = int64(sz)
		idx++
	}

	//Performance statistics initilization.
	ps, _ := NewPerfStats()

	//Start uploading objects
	var wg sync.WaitGroup
	wg.Add(len(sa))
	defer wg.Wait()
	for _, sz := range sa {
		log.Printf("Uploading object [%d]*%d.", sz, svr.Count)
		uploadRandomObj(&wg, ps, s3svc, bucket, sz, svr.Count)
	}

	ps.Shutdown()

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
func uploadRandomObj(wg *sync.WaitGroup, ps *PerfStats, svc *s3.S3, bucket *string, size int64, count int) error {
	defer wg.Done()

	for i := 0; i < int(count); i++ {
		obj, err := NewRandomObject("", "Test", size)
		if err != nil {
			panic("Failed to new random object!")
		}

		params := &s3.PutObjectInput{
			Bucket: aws.String(*bucket),
			Key:    aws.String(obj.Key),
			Body:   obj,
		}

		t := time.Now()

		_, err2 := svc.PutObject(params)
		if err2 != nil {
			panic("Failed to upload object" + err2.Error())
		}

		s := fmt.Sprintf("[PUT %d]", size)
		ps.PostSample(s, size, (time.Since(t)).Nanoseconds()/1000000, size, 0)
	}

	return nil
}
