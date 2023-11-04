package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"net/http"
	_ "net/http/pprof" // Note the underscore, which means import for side effects

	"github.com/klauspost/compress/zstd"
)

const (
	MAX_PACKET_SIZE = 1024 // Same as in sender
	ACK_MSG         = "ACK"
	maxBatchSize    = 1400 // to stay under the typical MTU
	maxBatchPackets = 10   // max number of packets before sending
)

// Packet structure must be the same as in sender
type Packet struct {
	SequenceNumber int
	Data           []byte
}

// decryptData decrypts the data using AES
func decryptData(encrypted []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(encrypted) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}

	// The IV is the first block of bytes
	iv := encrypted[:aes.BlockSize]
	encrypted = encrypted[aes.BlockSize:]

	// CFB decrypter
	stream := cipher.NewCFBDecrypter(block, iv)
	decrypted := make([]byte, len(encrypted))
	stream.XORKeyStream(decrypted, encrypted)

	return decrypted, nil
}

// decompressWithZstd decompresses data using zstd
func decompressWithZstd(data []byte) ([]byte, error) {
	decoder, err := zstd.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer decoder.Close()

	decompressed, err := io.ReadAll(decoder)
	if err != nil {
		return nil, err
	}

	return decompressed, nil
}

// PacketHeaderSize is the size of the packet header that contains the packet size.
const PacketHeaderSize = 4 // Let's assume we're using 4 bytes for the header.

// deserializeBatch takes a slice of bytes that represents a batch of serialized packets.
// It returns the first deserialized Packet, the rest of the batch, and any error encountered.
func deserializeBatch(batchData []byte) (Packet, []byte, error) {
	// Ensure there's enough data for the packet size header.
	if len(batchData) < PacketHeaderSize {
		return Packet{}, nil, errors.New("batch data is too small for a packet header")
	}

	// Read the size of the first packet.
	packetSize := binary.BigEndian.Uint32(batchData[:PacketHeaderSize])
	expectedPacketEnd := PacketHeaderSize + int(packetSize)

	// log.Printf("Raw packet size header bytes: %v", batchData[:PacketHeaderSize])
	// log.Printf("Expected packet size: %d bytes", packetSize)

	// Ensure the batch contains the full packet.
	if len(batchData) < expectedPacketEnd {
		return Packet{}, nil, errors.New("batch data does not contain the full packet")
	}

	// Extract the packet data.
	packetData := batchData[PacketHeaderSize:expectedPacketEnd]

	// Deserialize the packet.
	var packet Packet
	if err := deserialize(packetData, &packet); err != nil {
		return Packet{}, nil, fmt.Errorf("failed to deserialize packet: %w", err)
	}

	// Return the deserialized packet and the remaining batch data.
	return packet, batchData[expectedPacketEnd:], nil
}

// deserialize is a helper function that deserializes a single packet.
func deserialize(data []byte, pkt *Packet) error {
	// Here you would deserialize 'data' into 'pkt'.
	// For example, using encoding/gob, json, or any other serialization method you have chosen.
	return json.Unmarshal(data, pkt)
}

// receiveFile listens for incoming file data over UDP
func receiveFile(conn *net.UDPConn, destinationFile string, encryptionKey []byte) error {
	file, err := os.Create(destinationFile)
	if err != nil {
		return err
	}
	defer file.Close()

	var batchData []byte // Assuming batchData is accumulated over multiple packets.

	for {
		buffer := make([]byte, 8192) // Adjust buffer size as necessary.
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			return err
		}

		// log.Printf("Received %d bytes from %v", n, addr)

		// Accumulate batch data
		batchData = append(batchData, buffer[:n]...)

		// Try to deserialize the batch, handle incomplete packets without panicking
		packet, restOfBatch, err := deserializeBatch(batchData)
		if err != nil {
			if err.Error() == "batch data does not contain the full packet" {
				log.Println("Warning: Received incomplete packet. Waiting for more data.")
				continue // Wait for more data to arrive
			}
			// For other errors, you might want to handle them differently
			return fmt.Errorf("failed to deserialize batch: %v", err)
		}

		// TODO: Process the single packet here, such as decompression, decryption, writing to file, etc.
		// Decompress the data
		decompressedData, err := decompressWithZstd(packet.Data)
		if err != nil {
			return err
		}

		// Decrypt the data
		decryptedData, err := decryptData(decompressedData, encryptionKey)
		if err != nil {
			return err
		}

		// Write the decrypted data to the file
		_, err = file.Write(decryptedData)
		if err != nil {
			return err
		}

		// Send back an acknowledgment for the packet processed
		_, err = conn.WriteToUDP([]byte(ACK_MSG), addr)
		if err != nil {
			return err
		}

		// Prepare for next iteration with remaining batch data
		batchData = restOfBatch
	}

	return nil
}

func main() {
	go func() {
		// Start a HTTP server that will serve the pprof endpoints.
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	addr := net.UDPAddr{
		Port: 12345,
		IP:   net.ParseIP("0.0.0.0"),
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatalf("Error setting up UDP listener: %v", err)
	}
	defer conn.Close()

	encryptionKey := []byte("12345678901234567890123456789012") // AES key

	err = receiveFile(conn, "output.txt", encryptionKey)
	if err != nil {
		log.Fatalf("Error receiving file: %v", err)
	}
}
