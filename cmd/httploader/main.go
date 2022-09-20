package main

import (
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/danilovkiri/dk_go_url_shortener/internal/api/rest/modeldto"
)

func randStringBytes(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func main() {
	a := flag.String("a", "http://localhost:8080", "Server address")
	flag.Parse()
	address := *a

	const postRegular = "/"
	const postJSON = "/api/shorten"
	const postBatchJSON = "/api/shorten/batch"
	const getRegular = "/"
	const getAllByUserID = "/api/user/urls"
	const deleteBatch = "/api/user/urls"
	const ping = "/ping"
	const iterations = 20

	client := resty.New()
	client.SetRedirectPolicy(resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}))

	// Performing ping loading
	log.Println("Performing ping loading")
	for i := 0; i < iterations; i++ {
		_, err := client.R().Get(address + ping)
		if err != nil {
			log.Fatal(err)
		}
	}
	time.Sleep(1 * time.Second)

	// Performing postRegular loading
	log.Println("Performing postRegular loading")
	var sURLs []string
	for i := 0; i < iterations; i++ {
		payload := strings.NewReader("https://www." + randStringBytes(10) + ".com")
		res, err := client.R().SetBody(payload).Post(address + postRegular)
		if err != nil {
			log.Fatal(err)
		}
		if res.StatusCode() == 201 {
			sURLslice := strings.Split(string(res.Body()), "/")
			sURL := sURLslice[len(sURLslice)-1]
			sURLs = append(sURLs, sURL)
		}
	}
	log.Println(sURLs)
	time.Sleep(1 * time.Second)

	// Performing postJSON loading
	log.Println("Performing postJSON loading")
	for i := 0; i < iterations; i++ {
		URL := modeldto.RequestURL{
			URL: "https://www." + randStringBytes(10) + ".com",
		}
		reqBody, err := json.Marshal(URL)
		if err != nil {
			log.Fatal(err)
		}
		payload := strings.NewReader(string(reqBody))
		_, err = client.R().SetBody(payload).Post(address + postJSON)
		if err != nil {
			log.Fatal(err)
		}
	}
	time.Sleep(1 * time.Second)

	// Performing postBatchJSON loading
	log.Println("Performing postBatchJSON loading")
	for i := 0; i < iterations; i++ {
		batch := []modeldto.RequestBatchURL{
			{
				CorrelationID: "test1",
				URL:           "https://www." + randStringBytes(10) + ".com",
			},
			{
				CorrelationID: "test2",
				URL:           "https://www." + randStringBytes(10) + ".com",
			},
		}
		reqBody, err := json.Marshal(batch)
		if err != nil {
			log.Fatal(err)
		}
		payload := strings.NewReader(string(reqBody))
		_, err = client.R().SetBody(payload).Post(address + postBatchJSON)
		if err != nil {
			log.Fatal(err)
		}
	}
	time.Sleep(1 * time.Second)

	// Performing getRegular loading
	log.Println("Performing getRegular loading")
	for i := 0; i < iterations; i++ {
		_, err := client.R().Get(address + getRegular + sURLs[i])
		if err != nil {
			log.Fatal(err)
		}
	}
	time.Sleep(1 * time.Second)

	// Performing getAllByUserID loading
	log.Println("Performing getAllByUserID loading")
	for i := 0; i < iterations; i++ {
		res, err := client.R().Get(address + getAllByUserID)
		if err != nil {
			log.Fatal(err)
		}
		if res.StatusCode() == 200 {
			log.Println("Iteration", i, string(res.Body()))
		}
	}
	time.Sleep(1 * time.Second)

	// Performing deleteBatch loading
	log.Println("Performing deleteBatch loading")
	for i := 0; i < iterations; i++ {
		reqBody, err := json.Marshal([]string{sURLs[i]})
		if err != nil {
			log.Fatal(err)
		}
		payload := strings.NewReader(string(reqBody))
		_, err = client.R().SetBody(payload).Delete(address + deleteBatch)
		if err != nil {
			log.Fatal(err)
		}
	}
	time.Sleep(15 * time.Second)
}
