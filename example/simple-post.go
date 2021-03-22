package main

import (
	"bytes"
	"flag"
	"fmt"
	"math/rand"
	"time"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	var prt = flag.Int("print", 2, "debug print")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	httpposturl := "http://" + os.Getenv("LB_ADDR") + "/rsrp"

	for i := 0; i < 1000; i++ {
		cell := rand.Uint32() % 256
		user := rand.Uint32() % 1000
		num := (rand.Uint32() % 99) + 1

		var postString = fmt.Sprintf("{\"cellname\":\"%d\",\"username\":\"%d\",\"rsrp\":%d}", cell, user, num)
		var jsonData = []byte(postString)

		request, error := http.NewRequest("POST", httpposturl, bytes.NewBuffer(jsonData))
		request.Header.Set("Content-Type", "application/json; charset=UTF-8")

		client := &http.Client{}
		response, error := client.Do(request)
		if error != nil {
			panic(error)
		}

		if *prt < 2 {
			fmt.Println("response Status:", response.Status)
		}

		if *prt == 0 || response.StatusCode != 200 {
			fmt.Println("HTTP JSON POST URL:", httpposturl)
			fmt.Printf("cell = %d, user = %d, num = %d\n", cell, user, num)
			fmt.Println(postString)
			fmt.Println("response Headers:", response.Header)
			body, _ := ioutil.ReadAll(response.Body)
			fmt.Println("response Body:", string(body))
		}

		response.Body.Close()
	}
}
