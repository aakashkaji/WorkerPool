package main

import (
	"fmt"
	"sync"
	"time"
	"go.mongodb.org/mongo-driver/mongo"
	"io/ioutil"
	"net/http"
	b64 "encoding/base64"

	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"context"
)

type Job struct {
	ID int
	base_string       string
}
type Result struct {
	job         Job
	HtmlData  string
}

type Data1 struct {
	ID int `json:"id"`
	HtmlData string `json:"html_data"`
}

var jobs = make(chan Job, 500)
var results = make(chan Result, 500)


func recovery_f() {

	if r := recover(); r != nil {fmt.Println("some error has occured_______________________________________________",r)}
}

func html_hit(url string)string {

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("authority", "www.aireaa.com")
	req.Header.Add("pragma", "no-cache")
	req.Header.Add("cache-control", "no-cache")
	req.Header.Add("accept", "*/*")
	req.Header.Add("x-requested-with", "XMLHttpRequest")
	req.Header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_2) AppleWebKit/537.36 (KHTML like Gecko) Chrome/79.0.3945.88 Safari/537.36")
	req.Header.Add("sec-fetch-site", "same-origin")
	req.Header.Add("sec-fetch-mode", "cors")
	req.Header.Add("referer", "https://www.aireaa.com/agent.php")
	req.Header.Add("accept-encoding", "gzip deflate")
	req.Header.Add("accept-language", "en-GBen-US;q=0.9en;q=0.8")
	req.Header.Add("cookie", "__cfduid=db40d34e371afc25d27c90e5ab8bb075e1578401240; PHPSESSID=f1dfee6d9aa3e8e8bbfb64668e1c3fe2; _ga=GA1.2.1115224252.1578401246; _gid=GA1.2.409544342.1578401246,__cfduid=db40d34e371afc25d27c90e5ab8bb075e1578401240; PHPSESSID=f1dfee6d9aa3e8e8bbfb64668e1c3fe2; _ga=GA1.2.1115224252.1578401246; _gid=GA1.2.409544342.1578401246; __cfduid=d0e9aa2c9461ad2dd5666cc8eb50b12361578420415; PHPSESSID=2ceb85f9a91c99cb8d0124f78d08e644")
	req.Header.Add("Host", "www.aireaa.com")
	req.Header.Add("Connection", "keep-alive")

	res, err:= http.DefaultClient.Do(req)
	//res, err := c.Do(req)
	if err != nil {panic(err)}


	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	return fmt.Sprintf("%s", string(body))

}

func hit(base64 string) string{

	defer recovery_f()

	url := fmt.Sprintf("https://www.aireaa.com/agent_detail.php?id=%s",base64)



	return html_hit(url)

}
func worker(wg *sync.WaitGroup) {
	for job := range jobs {
		output := Result{job, hit(job.base_string)}
		results <- output
	}
	wg.Done()
}
func createWorkerPool(noOfWorkers int) {
	var wg sync.WaitGroup
	for i := 0; i < noOfWorkers; i++ {
		wg.Add(1)
		go worker(&wg)
	}
	wg.Wait()
	close(results)
}
func allocate(noOfJobs int) {
	for i := 1; i < noOfJobs; i++ {
		fmt.Println("________________________________________no of Jobs", i)
		c := fmt.Sprintf("%d", i)
		sEnc := b64.StdEncoding.EncodeToString([]byte(c))
		job := Job{i,sEnc}
		jobs <- job
	}
	close(jobs)
}
func result(done chan bool) {
	co := mongoConnection1()

	collection := co.Database("aakash").Collection("mydata")
	for result := range results {

		x := Data1{result.job.ID, result.HtmlData}

		insert, err1 := collection.InsertOne(context.TODO(), x)
		fmt.Println(insert)
		if err1 != nil {
			panic(err1)
		}
	}

	done <- true
}

func mongoConnection1() *mongo.Client {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	return client

}

func main() {


	startTime := time.Now()
	noOfJobs := 92212
	go allocate(noOfJobs)
	done := make(chan bool)
	go result(done)
	noOfWorkers := 50
	createWorkerPool(noOfWorkers)
	fmt.Println(<-done)
	endTime := time.Now()
	diff := endTime.Sub(startTime)
	fmt.Println("total time taken ", diff.Seconds(), "seconds")
}
