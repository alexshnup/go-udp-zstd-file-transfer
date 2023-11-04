# UDP File Transfer System

This repository contains two applications that collectively facilitate the secure and efficient transfer of files over UDP. The `sender` application encrypts and compresses files, then sends them to the `receiver` application, which decrypts and decompresses the received files.

## Features

- **AES Encryption**: Secures file contents during transfer.
- **Zstandard Compression**: Minimizes file size for faster transmission.
- **UDP Protocol**: Provides quick transfer speeds compared to TCP.

## Prerequisites

- Go programming language (version 1.15 or higher recommended).
- zstd library for Go.

## Installation

To get started, clone this repository to your local machine:

```bash
git clone https://github.com/alexshnup/go-udp-zstd-file-transfer.git
cd go-udp-zstd-file-transfer
go mod tidy
```

## Configuration

Before running the applications, you must configure both with the same AES encryption key. This key must be 16, 24, or 32 bytes in length corresponding to AES-128, AES-192, or AES-256 encryption, respectively. Modify the encryptionKey variable in the main.go file for both sender and receiver to ensure they match.

## Building the Applications
```bash
# Build receiver
cd receiver
go build

# Build sender
cd ../sender
go build
```

## Usage

### Receiver
To start the receiver, which will listen for incoming files, run:
```bash
./receiver
```
The receiver will be listening on 0.0.0.0 at port 12345. Ensure this port is open and accessible where the receiver is running.

### Sender
To send a file, run the sender application with the filename as an argument:
```bash
./sender path/to/your/file
```
Ensure the receiver is running before you start the sender. The sender will connect to the receiver's IP address and port (default is localhost:12345) to initiate the file transfer.



## Network Configuration
By default, the sender is configured to send files to localhost:12345. To send files across different machines, update the serverAddr variable in the sender's main.go to match the receiver's IP address and port.

## Limitations
The error handling is currently minimal; network timeouts and errors will lead to retries without exponential backoff.
The applications are pre-configured for use within a local network. For internet transfers, consider implementing proper NAT traversal techniques.
Contributing
Contributions are welcome! Please fork the project, include your changes, and submit a pull request for review.

## License
Specify your license here, or state that the project is in the public domain.

## Acknowledgements
The Go community for the comprehensive libraries and support.
Klauspost for the zstd compression library.
For a deeper dive into the codebase, please review the comments provided within each Go source file.

