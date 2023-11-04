package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/klauspost/compress/zstd"

	"net/http"
	_ "net/http/pprof" // Note the underscore, which means import for side effects
)

const (
	MAX_PACKET_SIZE = 1024 // Adjust based on your network's MTU
	ACK_MSG         = "ACK"
	maxBatchSize    = 1400 // to stay under the typical MTU
	maxBatchPackets = 10   // max number of packets before sending
)

// Packet represents the data structure to send over the network
type Packet struct {
	SequenceNumber int
	Data           []byte
}

// type PacketHeader struct {
// 	Size int
// }

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
	encoder, err := zstd.NewWriter(nil)
	if err != nil {
		return nil, err
	}
	defer encoder.Close()

	return encoder.EncodeAll(data, nil), nil
}

func serialize(packet Packet) ([]byte, error) {
	var buf bytes.Buffer

	// Serialize packet data using JSON
	packetData, err := json.Marshal(packet)
	if err != nil {
		return nil, err
	}

	// Log the size of the JSON serialized data
	// log.Printf("JSON serialized data size: %d bytes\n", len(packetData))

	// Write the size of the JSON serialized packet data in big-endian format
	packetSize := uint32(len(packetData))
	if err := binary.Write(&buf, binary.BigEndian, packetSize); err != nil {
		return nil, err
	}

	// Log the size of the packet header
	// log.Printf("Packet header size: %d bytes\n", 4) // Size of uint32

	// Write the actual packet data
	if _, err := buf.Write(packetData); err != nil {
		return nil, err
	}

	// Log the total size of the serialized packet
	// log.Printf("Total serialized packet size: %d bytes\n", buf.Len())

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
	var batch []byte
	var batchCount int

	sendBatch := func() error {
		if len(batch) > 0 {
			_, err := conn.WriteToUDP(batch, addr)
			batch = batch[:0] // Reset the batch
			batchCount = 0
			return err
		}
		return nil
	}

	for {
		bytesRead, err := file.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break // End of file reached
			}
			return err
		}
		fileChunk := buffer[:bytesRead]

		// Logging the size of the file chunk before encryption
		// log.Printf("File chunk size (before encryption): %d bytes\n", len(fileChunk))

		// Encrypt and then compress the chunk of data
		encryptedChunk, err := encryptData(fileChunk, encryptionKey)
		if err != nil {
			return err
		}

		// Logging the size after encryption and before compression
		// log.Printf("Encrypted chunk size (before compression): %d bytes\n", len(encryptedChunk))

		compressedChunk, err := compressWithZstd(encryptedChunk)
		if err != nil {
			return err
		}

		// Logging the size after compression
		// log.Printf("Compressed chunk size: %d bytes\n", len(compressedChunk))

		// Create a packet and serialize it
		packet := Packet{
			SequenceNumber: sequenceNumber,
			Data:           compressedChunk,
		}
		serializedPacket, err := serialize(packet)
		if err != nil {
			return err
		}

		// Logging the size of the serialized packet
		// log.Printf("Serialized packet size: %d bytes\n", len(serializedPacket))

		// Accumulate the packet in the batch
		if len(batch)+len(serializedPacket) > maxBatchSize || batchCount >= maxBatchPackets {
			if err := sendBatch(); err != nil {
				return err
			}
		}
		batch = append(batch, serializedPacket...)
		batchCount++

		// Wait for acknowledgment before sending the next packet
		// ... (acknowledgment code would need to be adjusted to handle batches)

		// Increment the sequence number for the next packet
		sequenceNumber++
	}

	// Send any remaining batched data
	return sendBatch()
}

func main() {
	// save current time for calculating total time taken
	start := time.Now()
	go func() {
		// Start a HTTP server that will serve the pprof endpoints.
		// Do not expose this in a production environment; it's only for profiling purposes.
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
	encryptionKey := []byte("12345678901234567890123456789012") // 32 bytes for AES-256

	if len(os.Args) < 2 {
		panic("Please provide a filename to send")
	}
	fienameFromArgs := os.Args[1]

	err = sendFile(conn, udpAddr, fienameFromArgs, encryptionKey)
	if err != nil {
		panic(err)
	}

	// Calculate total time taken
	elapsed := time.Since(start)
	log.Printf("Total time taken: %s", elapsed)
}
