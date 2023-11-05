package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"os"
	"sync"

	"github.com/klauspost/compress/zstd"
)

const (
	MAX_PACKET_SIZE = 1024
	ACK_MSG         = "ACK"
	SHARD_SIZE      = 65536 // Must be the same as in the sender
	START_PORT      = 30000 // Must be the same as in the sender
	TOTAL_SHARDS    = 2     // Adjust based on expected number of shards
)

type Packet struct {
	SequenceNumber int
	Data           []byte
}

func decryptData(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	decrypted := make([]byte, len(data))
	stream.XORKeyStream(decrypted, data)

	return decrypted, nil
}

func decompressWithZstd(data []byte) ([]byte, error) {
	decoder, err := zstd.NewReader(nil)
	if err != nil {
		return nil, err
	}
	defer decoder.Close()

	return decoder.DecodeAll(data, nil)
}

func deserialize(data []byte) (Packet, error) {
	var packet Packet
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	err := decoder.Decode(&packet)
	return packet, err
}

// func receiveShard(port int, wg *sync.WaitGroup, encryptionKey []byte, shardDataChan chan<- []byte) {
// 	defer wg.Done()

// 	addr := net.UDPAddr{
// 		Port: port,
// 		IP:   net.ParseIP("0.0.0.0"),
// 	}
// 	conn, err := net.ListenUDP("udp", &addr)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer conn.Close()

// 	var fileData []byte
// 	for {
// 		buffer := make([]byte, MAX_PACKET_SIZE)
// 		n, _, err := conn.ReadFromUDP(buffer)
// 		if err != nil {
// 			if err == io.EOF {
// 				break // End of shard
// 			}
// 			panic(err) // or handle error differently
// 		}

// 		packet, _ := deserialize(buffer[:n])
// 		decryptedChunk, _ := decryptData(packet.Data, encryptionKey)
// 		decompressedChunk, _ := decompressWithZstd(decryptedChunk)

// 		fileData = append(fileData, decompressedChunk...)

// 		_, err = conn.Write([]byte(ACK_MSG))
// 		if err != nil {
// 			panic(err) // or handle error differently
// 		}
// 	}

// 	shardDataChan <- fileData
// }

func receiveShard(port int, wg *sync.WaitGroup, encryptionKey []byte, shardDataChan chan<- []byte) {
	defer wg.Done()

	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("0.0.0.0"),
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	var fileData []byte
	for {
		buffer := make([]byte, MAX_PACKET_SIZE)
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if err == io.EOF {
				break // End of shard
			}
			panic(err) // or handle error differently
		}

		// Log the size and contents of the received data
		// fmt.Printf("Received packet: size=%d, contents=%s\n", n, buffer[:n])

		// if n <= aes.BlockSize {
		// 	fmt.Printf("Received data is too short: %d bytes\n", n)
		// 	continue // Skip this packet
		// }

		packet, err := deserialize(buffer[:n])
		if err != nil {
			fmt.Println("Failed to deserialize packet:", err)
			continue // Skip this packet
		}

		// After deserializing
		fmt.Printf("Deserialized packet: %+v\n", packet)
		// After decompressing the data
		fmt.Printf("SequenceNumber: %d\n", packet.SequenceNumber)
		// fmt.Printf("Data: %s\n", packet.Data)

		// decryptedChunk, err := decryptData(packet.Data, encryptionKey)
		// if err != nil {
		// 	fmt.Println("Failed to decrypt data:", err)
		// 	continue // Skip this packet
		// }

		// decompressedChunk, err := decompressWithZstd(decryptedChunk)
		// if err != nil {
		// 	fmt.Println("Failed to decompress data:", err)
		// 	continue // Skip this packet
		// }

		// fileData = append(fileData, decompressedChunk...)
		fileData = append(fileData, packet.Data...)

		_, err = conn.Write([]byte(ACK_MSG))
		if err != nil {
			// panic(err) // or handle error differently
			fmt.Println("Failed to send ACK:", err)
		}
	}

	shardDataChan <- fileData
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

	var wg sync.WaitGroup
	encryptionKey := []byte("1234567890123456")
	shardDataChan := make(chan []byte, TOTAL_SHARDS)

	for i := 0; i < TOTAL_SHARDS; i++ {
		wg.Add(1)
		go receiveShard(START_PORT+i, &wg, encryptionKey, shardDataChan)
	}

	go func() {
		wg.Wait()
		close(shardDataChan)
	}()

	outputFile, err := os.Create("received_file")
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	for shardData := range shardDataChan {
		outputFile.Write(shardData)
	}
}
