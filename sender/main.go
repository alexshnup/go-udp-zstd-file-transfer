package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	_ "net/http/pprof" // Import for side effects

	"github.com/klauspost/compress/zstd"
)

const (
	MAX_PACKET_SIZE = 1024 // Adjust based on your network's MTU
	ACK_MSG         = "ACK"
)

// Packet represents the data structure to send over the network
type Packet struct {
	SequenceNumber int
	Data           []byte
}

func encryptData(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// The IV needs to be unique, but not secret
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// CFB encrypter
	stream := cipher.NewCFBEncrypter(block, iv)
	encrypted := make([]byte, len(data))
	stream.XORKeyStream(encrypted, data)

	return append(iv, encrypted...), nil
}

func compressWithZstd(data []byte) ([]byte, error) {
	// fast compression level
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

func sendFile(conn *net.UDPConn, addr *net.UDPAddr, filename string, encryptionKey []byte) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	buffer := make([]byte, MAX_PACKET_SIZE)
	sequenceNumber := 1

	for {
		bytesRead, err := file.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break // End of file reached
			}
			return err
		}
		fileChunk := buffer[:bytesRead]

		// Encrypt and then compress the chunk of data
		encryptedChunk, err := encryptData(fileChunk, encryptionKey)
		if err != nil {
			return err
		}
		compressedChunk, err := compressWithZstd(encryptedChunk)
		if err != nil {
			return err
		}

		// Create a packet and serialize it
		packet := Packet{
			SequenceNumber: sequenceNumber,
			Data:           compressedChunk,
		}
		serializedPacket, err := serialize(packet)
		if err != nil {
			return err
		}

		// Send the serialized packet
		_, err = conn.WriteToUDP(serializedPacket, addr)
		if err != nil {
			return err
		}

		// Wait for acknowledgment before sending the next packet
		ackBuffer := make([]byte, len(ACK_MSG)+10)
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		_, _, err = conn.ReadFromUDP(ackBuffer)
		if err != nil {
			// Handle the timeout or error, possibly resending the packet
			// For the example, let's just retry without backoff (not recommended for production)
			continue
		}

		// Increment the sequence number for the next packet
		sequenceNumber++
	}

	return nil
}

func main() {
	//pprof
	go func() {
		// Start a HTTP server that will serve the pprof endpoints.
		log.Println(http.ListenAndServe("localhost:6061", nil))
	}()

	serverAddr := "localhost:12345" // The address of the receiver
	udpAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		panic(err)
	}

	localAddr := net.UDPAddr{
		Port: 0, // Let the system assign a port
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", &localAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Use the same key as the receiver for AES encryption
	encryptionKey := []byte("1234567890123456") // 32 bytes for AES-256

	if len(os.Args) < 2 {
		panic("Please provide a filename to send")
	}
	fienameFromArgs := os.Args[1]

	//save current time to calculate time taken
	start := time.Now()

	err = sendFile(conn, udpAddr, fienameFromArgs, encryptionKey)
	if err != nil {
		panic(err)
	}

	//calculate time taken
	elapsed := time.Since(start)
	fmt.Println("Time taken: ", elapsed)
}
