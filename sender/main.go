package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

const (
	port         = 30000
	maxChunkSize = 1024 // Maximum size of each chunk
)

var serverIP string

func sendChunk(serverAddr string, chunk []byte, sequence int, totalChans int) {
	conn, err := net.Dial("udp", serverAddr)
	if err != nil {
		fmt.Println("Error dialing:", err)
		return
	}
	defer conn.Close()

	// Prepare a 4-byte buffer to hold the sequence number
	seqBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(seqBuf, uint32(sequence)) // Encode sequence number as 4 bytes

	// Append the sequence buffer and the chunk data
	data := append(seqBuf, chunk...)

	_, err = conn.Write(data)
	if err != nil {
		fmt.Println("Error sending data:", err)
		return
	}

	// if sequence == totalChans-1 {
	// 	time.Sleep(10 * time.Second)
	// }
	// fmt.Printf("Sent chunk %d\n", sequence)
}

func main() {

	//get ip from hostname
	ip, err := net.LookupIP("receiver")
	if err != nil {
		panic(err)
	}
	fmt.Printf("IP: %v\n", ip[0])
	serverIP = fmt.Sprintf("%v", ip[0])

	if len(os.Args) < 2 {
		panic("Please provide a filename to send")
	}
	filename := os.Args[1]

	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()
	totalChunks := int(fileSize / maxChunkSize)
	if fileSize%maxChunkSize != 0 {
		totalChunks++
	}

	log.Printf("Total chunks: %d\n", totalChunks)

	serverAddr := fmt.Sprintf("%s:%d", serverIP, port)
	conn, err := net.Dial("udp", serverAddr)
	if err != nil {
		fmt.Println("Error dialing:", err)
		return
	}
	_, err = conn.Write([]byte(fmt.Sprintf("total:%d", totalChunks)))
	if err != nil {
		fmt.Println("Error sending control message:", err)
		return
	}
	conn.Close()

	time.Sleep(100 * time.Millisecond)

	buffer := make([]byte, maxChunkSize)
	sequence := 0
	for {
		bytesRead, err := file.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break // End of file reached
			}
			fmt.Println("Error reading from file:", err)
			return
		}

		// log.Printf("\x1b[33m Sending chunk \x1b[35m  %d \x1b[0m of %d\n", sequence+1, totalChunks)

		if totalChunks > 10 && sequence%(totalChunks/10) == 0 {
			log.Printf("\x1b[33m Sending chunk \x1b[35m  %d \x1b[0m of %d\n", sequence+1, totalChunks)
		}

		// if sequence > totalChunks-10 {
		// 	log.Printf("\x1b[33m Sending chunk \x1b[35m  %d \x1b[0m of %d\n", sequence+1, totalChunks)
		// }
		outData := make([]byte, maxChunkSize)
		copy(outData, buffer[:bytesRead])
		// outData := append([]byte(fmt.Sprintf("%d:", sequence)), buffer[:bytesRead]...)

		// go sendChunk(serverAddr, buffer[:bytesRead], sequence, totalChunks)
		sendChunk(serverAddr, outData, sequence, totalChunks)
		// time.Sleep(1 * time.Microsecond)
		// time.Sleep(1 * time.Millisecond)
		// time.Sleep(1 * time.Second)
		sequence++

	}
	time.Sleep(2 * time.Second)

	// fmt.Println("All chunks sent")

}
