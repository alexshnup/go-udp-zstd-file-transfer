package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/klauspost/compress/zstd"
)

const (
	MAX_PACKET_SIZE = 1024
	ACK_MSG         = "ACK"
	// SHARD_SIZE      = 65536     // Size of each shard in bytes
	SHARD_SIZE = 2     // Size of each shard in bytes
	START_PORT = 30000 // Starting port for sharding
)

type Packet struct {
	SequenceNumber int
	Data           []byte
}

func encryptData(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	encrypted := make([]byte, len(data))
	stream.XORKeyStream(encrypted, data)

	return append(iv, encrypted...), nil
}

func compressWithZstd(data []byte) ([]byte, error) {
	encoder, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedFastest))
	if err != nil {
		return nil, err
	}
	defer encoder.Close()

	return encoder.EncodeAll(data, nil), nil
}

func serialize(packet Packet) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(packet)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func sendShard(shard []byte, port int, wg *sync.WaitGroup, encryptionKey []byte) {
	defer wg.Done()

	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("172.19.0.2"), // Adjust as necessary
	}
	conn, err := net.DialUDP("udp", nil, &addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	sequenceNumber := 1
	for i := 0; i < len(shard); i += MAX_PACKET_SIZE {
		end := i + MAX_PACKET_SIZE
		if end > len(shard) {
			end = len(shard)
		}

		fmt.Printf("Original data before encryption: %s\n", shard[i:end])

		// encryptedChunk, _ := encryptData(shard[i:end], encryptionKey)
		// compressedChunk, _ := compressWithZstd(encryptedChunk)

		packet := Packet{
			SequenceNumber: sequenceNumber,
			// Data:           compressedChunk,
			Data: shard[i:end],
		}
		serializedPacket, _ := serialize(packet)

		_, err = conn.Write(serializedPacket)
		if err != nil {
			panic(err)
		}

		// Log the size and contents of the packet before sending
		// fmt.Printf("Sending packet: size=%d, contents=%s\n", len(serializedPacket), serializedPacket)
		fmt.Printf("Sending packet: size=%d, \n", len(serializedPacket))

		ackBuffer := make([]byte, len(ACK_MSG)+10)
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		_, _, err = conn.ReadFromUDP(ackBuffer)
		if err != nil {
			continue // handle timeout or error, possibly resending the packet
		}

		sequenceNumber++
	}
}

func main() {
	if len(os.Args) < 2 {
		panic("Please provide a filename to send")
	}
	filename := os.Args[1]

	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()
	totalShards := (fileSize + SHARD_SIZE - 1) / SHARD_SIZE

	var wg sync.WaitGroup
	encryptionKey := []byte("1234567890123456")

	for i := int64(0); i < totalShards; i++ {
		start := i * SHARD_SIZE
		end := start + SHARD_SIZE
		if end > fileSize {
			end = fileSize
		}

		shard := make([]byte, end-start)
		file.ReadAt(shard, start)

		wg.Add(1)
		go sendShard(shard, START_PORT+int(i), &wg, encryptionKey)
	}

	wg.Wait()
}
