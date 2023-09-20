package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

func main() {
	connection, err := net.Dial("tcp", "localhost:4444")
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Println("Starting client")

	totalRequest := 10_000
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		i := 0
		for ; i < totalRequest; i++ {
			n, err := connection.Write([]byte("hey "))
			if err != nil {
				log.Println("error sending message", err)
				break
			}

			if n == 0 {
				log.Println("wrote only 0 bytes, something is weird")
			}
			time.Sleep(100 * time.Millisecond)
		}
		log.Println("Sent", i, "messages")
		wg.Done()
	}()

	go func() {
		totalResponses := 0
		for {
			buf := make([]byte, 1_500)
			n, err := connection.Read(buf)
			if err != nil {
				log.Println("error reading from connection", err)
				break
			}

			if n == 0 {
				continue
			}

			log.Println("Bytes read", n)
			response := string(buf[:n])
			response = strings.TrimSpace(response)
			if response == "" {
				continue
			}
			totalResponses += len(strings.Split(response, " "))
		}
		log.Println("Total Responses", totalResponses)
		wg.Done()
	}()

	wg.Wait()
}
