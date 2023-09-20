package main

import (
	"bytes"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {

	addr := net.TCPAddr{Port: 4444}
	listener, err := net.ListenTCP("tcp", &addr)
	if err != nil {
		log.Fatal(err)
	}

	sigChan := make(chan os.Signal, 1)
	shuttingDown := make(chan struct{}, 1)

	go func() {
		log.Println("listening for new connections")
		for {
			select {
			case <-shuttingDown:
				log.Println("Shutting down listener")
				_ = listener.Close()
				return
			default:
				conn, err := listener.AcceptTCP()
				if err != nil {
					log.Println(err)
					return
				}

				go handleConnection(conn, shuttingDown)
			}
		}
	}()

	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	for sig := range sigChan {
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			log.Println("syscall to shutdown")
			shuttingDown <- struct{}{}
			return
		}
	}

}

func handleConnection(conn *net.TCPConn, shuttingDown chan struct{}) {
	// if we are in different network from the client likely the max size of a
	// frame is 1.5kb, no need to make this bigger
	requestContent := make([]byte, 1500)
	responseBuffer := bytes.NewBuffer(make([]byte, 0, 1500))

	for {
		select {
		case <-shuttingDown:
			_ = conn.CloseRead()
			log.Println("connection closed")
			return
		default:
			n, err := conn.Read(requestContent)
			if err != nil {
				break
			}

			if n == 0 {
				continue
			}

			generateResponse(string(requestContent[:n]), responseBuffer)
			writeResponse(responseBuffer.Bytes(), conn)
			responseBuffer.Reset()
		}
	}
}

func generateResponse(request string, responseBuffer *bytes.Buffer) {
	yesResponse := "yes"
	newLineString := "\n"
	elements := strings.Split(request, " ")
	responses := make([]string, 0)
	for i := range elements {
		currentString := elements[i]
		if strings.TrimSpace(currentString) == "" {
			continue
		}
		responses = append(responses, yesResponse)
		if strings.HasSuffix(currentString, newLineString) {
			responses = append(responses, newLineString)
		}
	}

	responseBuffer.Write([]byte(strings.Join(responses, " ")))
}

func writeResponse(response []byte, conn *net.TCPConn) {
	_, err := conn.Write(response)
	if err != nil {
		log.Println("error writing response to client", err)
	}
}
