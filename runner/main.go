package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sync"
)

func main() {
	const numGoroutines = 100_000
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sendRequest(id)
		}(i)
	}

	wg.Wait()
	fmt.Println("All requests completed.")
}

func sendRequest(id int) {
	url := fmt.Sprintf("http://localhost:8080/v1/key%d", id)

	jsonBody := fmt.Sprintf(`value-%d`, id)

	req, err := http.NewRequest("PUT", url, bytes.NewBufferString(jsonBody))
	if err != nil {
		fmt.Printf("Goroutine %d: error creating request: %v\n", id, err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Goroutine %d: error sending request: %v\n", id, err)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Goroutine %d: error reading response: %v\n", id, err)
		return
	}

	fmt.Printf("Goroutine %d: Status: %s, Response: %s\n", id, resp.Status, string(respBody))
}
