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

const numWorkers = 4 // Number of worker goroutines for processing data

func processChunks(workerId int, chunksChan <-chan []byte, receivedChunks map[int][]byte, readyToWriteChan chan<- int, mutex *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done() // Ensure wg.Done() is called when the goroutine exits

	log.Printf("Worker %d started\n", workerId)
	for packet := range chunksChan {
		sequence := int(binary.BigEndian.Uint32(packet[:4]))
		chunkData := packet[4:]

		mutex.Lock()
		receivedChunks[sequence] = chunkData
		mutex.Unlock()
		// log.Printf("Worker %d received chunk %d len: %d %s\n", workerId, sequence+1, len(chunkData), chunkData[:10])

		readyToWriteChan <- sequence // Notify main goroutine
	}
	log.Printf("Worker %d finished\n", workerId)
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
	successReceivedChunksIds := make(map[int]bool)

	chunksChan := make(chan []byte, 100)    // Buffered channel
	readyToWriteChan := make(chan int, 100) // For notifying chunks ready to write

	var mutex sync.Mutex // Mutex for synchronizing access to receivedChunks
	var wg sync.WaitGroup

	// Start a single goroutine for processing chunks
	// go processChunks(chunksChan, readyToWriteChan, receivedChunks, &mutex)

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go processChunks(i, chunksChan, receivedChunks, readyToWriteChan, &mutex, &wg)

	}

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

	lowestUnwrittenChunk := 0
	totalChunksExpected := <-totalChunksChan
	fmt.Printf("Total chunks expected: %d\n", totalChunksExpected)

	for {
		select {
		// case sequence := <-readyToWriteChan:
		case <-readyToWriteChan:
			mutex.Lock()
			for {
				chunk, exists := receivedChunks[lowestUnwrittenChunk]
				if !exists {
					break // No contiguous chunk available
				}
				// log.Printf("Writing chunk %d\n", lowestUnwrittenChunk)
				writer.Write(chunk) // Write the chunk
				delete(receivedChunks, lowestUnwrittenChunk)
				successReceivedChunksIds[lowestUnwrittenChunk] = true
				lowestUnwrittenChunk++
			}
			mutex.Unlock()
			// if sequence == totalChunksExpected {
			// 	// Last chunk written
			// 	log.Printf("Last chunk written, lowestUnwrittenChunk=%d\n", lowestUnwrittenChunk)
			// 	goto Assemble
			// }
		case <-time.After(10 * time.Second):
			log.Println("Timeout reached, stopping reception")
			goto Assemble
		}

		if lowestUnwrittenChunk >= totalChunksExpected {
			log.Printf("All chunks written, lowestUnwrittenChunk=%d\n", lowestUnwrittenChunk)
			break // All chunks written
		}
	}

Assemble:
	// Make sure to close chunksChan when you're done with it
	// This will cause the worker goroutines to exit their loops
	close(chunksChan)
	log.Println("Assembling file...")
	wg.Wait() // Wait for all worker goroutines to finish
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
		if _, ok := successReceivedChunksIds[i]; ok {
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

	time.Sleep(1 * time.Second)

	// writer.Flush()
	// fmt.Println("File reassembled")
}
