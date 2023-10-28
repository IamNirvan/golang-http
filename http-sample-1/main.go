package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	TIMEOUT time.Duration = 5000
)

type Response struct {
	data []byte
	err  error
}

func main() {
	startTime := time.Now()

	// 1. create a context with a timeout for the http request
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*TIMEOUT)
	defer cancel()

	body, err := fetchData(ctx, "sam")
	if err != nil {
		log.Fatal("\ran error occurred when fetching data: " + err.Error())
	}

	formattedData := formatJSON(body)
	fmt.Println("\n Result: " + formattedData)
	fmt.Println("Took: " + time.Since(startTime).String())
}

func formatJSON(data []byte) string {
	var out bytes.Buffer
	err := json.Indent(&out, data, "", " ")

	if err != nil {
		fmt.Println(err)
	}

	d := out.Bytes()
	return string(d)
}

func fetchData(ctx context.Context, name string) ([]byte, error) {
	url := "https://api.genderize.io?name=" + name

	// 1. Make a new HTTP request
	request, requestErr := http.NewRequest("GET", url, nil)
	if requestErr != nil {
		return nil, requestErr
	}

	// 2. Make response channel to communicate result back to the main thread
	responseChan := make(chan Response)

	// 3. Execute the request to the third party API
	client := &http.Client{}
	go func() {
		body, err := makeHttpRequest(request, client)

		// 3.1 Write the data to the response channel
		responseChan <- Response{
			data: body,
			err:  err,
		}
	}()

	// 4. Run an infinite loop to check when a response is received or if the request timed out
	i := 0
	var loading string

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("request timed out")
		case response := <-responseChan:
			return response.data, response.err
		default:
			if i == 0 {
				loading = ".  "
			} else if i == 1 {
				loading = ".. "
			} else if i == 2 {
				loading = "..."
				i = -1
			}
			fmt.Printf("\rwaiting for response" + loading)
			time.Sleep(time.Millisecond * 130)
			i += 1
		}
	}
}

func makeHttpRequest(request *http.Request, client *http.Client) ([]byte, error) {
	// 1. Execute the request and fetch the response
	response, responseErr := client.Do(request)
	if responseErr != nil {
		return nil, responseErr
	}
	defer response.Body.Close()

	// 2. Read the contents of the response body
	body, bodyReadErr := io.ReadAll(response.Body)
	if bodyReadErr != nil {
		return nil, bodyReadErr
	}

	return body, nil
}
