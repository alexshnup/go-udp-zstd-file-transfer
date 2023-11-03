package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/gob"
	"errors"
	"io"
	"net"
	"os"

	"github.com/klauspost/compress/zstd"
)

const (
	MAX_PACKET_SIZE = 1024 // Same as in sender
	ACK_MSG         = "ACK"
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

// deserialize converts a byte slice into a Packet struct
func deserialize(data []byte) (Packet, error) {
	var packet Packet
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	err := decoder.Decode(&packet)
	if err != nil {
		return Packet{}, err
	}
	return packet, nil
}

// receiveFile listens for incoming file data over UDP
func receiveFile(conn *net.UDPConn, destinationFile string, encryptionKey []byte) error {
	file, err := os.Create(destinationFile)
	if err != nil {
		return err
	}
	defer file.Close()

	for {
		buffer := make([]byte, MAX_PACKET_SIZE+1024) // Slightly larger than expected to accommodate additional data
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			return err
		}
		if n == 0 {
			break // No more data to read
		}

		// Deserialize the packet
		packet, err := deserialize(buffer[:n])
		if err != nil {
			return err
		}

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

		// Send back an acknowledgment
		_, err = conn.WriteToUDP([]byte(ACK_MSG), addr)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	// Setup UDP listener
	addr := net.UDPAddr{
		Port: 12345, // Replace with your listening port
		IP:   net.ParseIP("0.0.0.0"),
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Use a key for AES decryption (must match the sender's key)
	encryptionKey := []byte("12345678901234567890123456789012") // Must be 16, 24, or 32 bytes long

	// Call receiveFile to start receiving data
	err = receiveFile(conn, "output.txt", encryptionKey)
	if err != nil {
		panic(err)
	}
}
