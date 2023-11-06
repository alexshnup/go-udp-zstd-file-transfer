package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	port       = 30000
	outputPath = "output_file" // Replace with your output file path
)

var mutex sync.Mutex // Mutex for synchronizing access to receivedChunks

const numWorkers = 4 // Number of worker goroutines for processing data

func processChunks(chunksChan <-chan []byte, readyToWriteChan chan<- int, receivedChunks map[int][]byte, mutex *sync.Mutex) {
	for packet := range chunksChan {
		sequence := int(binary.BigEndian.Uint32(packet[:4]))
		chunkData := packet[4:]

		mutex.Lock()
		receivedChunks[sequence] = chunkData
		mutex.Unlock()

		readyToWriteChan <- sequence // Notify that a new chunk is ready
	}
}

func main() {

	// get local ip address
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}
	var localIP string
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		// = ipv4
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			localIP = ipnet.IP.String()
			fmt.Println(localIP)
		}
	}

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println("Error resolving UDP address:", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error listening on UDP:", err)
		return
	}
	defer conn.Close()

	file, err := os.Create(outputPath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	totalChunksChan := make(chan int, 1) // Channel to receive totalChunksExpected
	receivedChunks := make(map[int][]byte)
	chunksChan := make(chan []byte, 100)    // Buffered channel
	readyToWriteChan := make(chan int, 100) // For notifying chunks ready to write

	// Start a single goroutine for processing chunks
	go processChunks(chunksChan, readyToWriteChan, receivedChunks, &mutex)

	// // Start worker goroutines
	// for i := 0; i < numWorkers; i++ {
	// 	fmt.Printf("start worker %d\n", i)
	// 	go processChunks(i, chunksChan, receivedChunks, &mutex)
	// }

	go func() {
		for {
			buffer := make([]byte, 2048) // Move buffer allocation here
			n, _, err := conn.ReadFromUDP(buffer)
			if err != nil {
				fmt.Println("Error reading UDP message:", err)
				return
			}

			if bytes.HasPrefix(buffer[:n], []byte("total:")) {
				totalChunksExpected, _ := strconv.Atoi(string(buffer[6:n]))
				totalChunksChan <- totalChunksExpected // Send totalChunksExpected to the main goroutine
				continue
			}

			chunksChan <- buffer[:n] // Send the data to be processed

		}
	}()

	totalChunksExpected := <-totalChunksChan
	fmt.Printf("Total chunks expected: %d\n", totalChunksExpected)

	lowestUnwrittenChunk := 0
	for {
		select {
		case sequence := <-readyToWriteChan:
			mutex.Lock()
			for {
				chunk, exists := receivedChunks[lowestUnwrittenChunk]
				if !exists {
					break // No contiguous chunk available
				}
				writer.Write(chunk) // Write the chunk
				delete(receivedChunks, lowestUnwrittenChunk)
				lowestUnwrittenChunk++
			}
			// fmt.Printf("Lowest unwritten chunk: %d sequence: %d\n", lowestUnwrittenChunk, sequence)
			mutex.Unlock()
			if sequence == totalChunksExpected-1 {
				fmt.Println("Last chunk received")
				goto Assemble
			}
		case <-time.After(10 * time.Second):
			log.Println("Timeout reached, stopping reception")
			goto Assemble
		}

		if lowestUnwrittenChunk >= totalChunksExpected {
			break // All chunks written
		}
	}

Assemble:
	writer.Flush()
	log.Println("File reassembled")

	// // print size of map

	// log.Println("Waiting for receiver to start...")
	// prevSize := 0
	// prevCount := 10
	// for {
	// 	log.Printf("\x1b[33m Size of map: \x1b[35m  %d \x1b[0m of %d\n", len(receivedChunks), totalChunksExpected)
	// 	if len(receivedChunks) == totalChunksExpected {
	// 		break
	// 	}
	// 	if len(receivedChunks) == prevSize {
	// 		prevCount--
	// 	}
	// 	if prevCount == 0 {
	// 		break
	// 	}
	// 	prevSize = len(receivedChunks)
	// 	time.Sleep(1 * time.Second)
	// }

	missedChunks := 0

	// Reassemble and write the file
	for i := 0; i < totalChunksExpected; i++ {
		if _, ok := receivedChunks[i]; ok {
			// if totalChunksExpected > 10 && i%(totalChunksExpected/10) == 0 {
			// 	// log.Printf("\x1b[33m Writing chunk: \x1b[35m  %d \x1b[0m of %d\n", i+1, totalChunksExpected)
			// }
			// writer.Write(chunk)
		} else {
			missedChunks++
			// fmt.Printf("Missing chunk: %d of %d\n", i+1, totalChunksExpected)
		}
	}
	fmt.Printf("Missed chunks: %d\n", missedChunks)

	// writer.Flush()
	// fmt.Println("File reassembled")
}
