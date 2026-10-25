package wire

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"

	"drift/internals/errs"
)

// Las conexiones se crearan mendiante un chanel que incluira id de peer y otros datos, de momento escucha de cualquiera

func Listen(maxConns, maxMsgStreamSize, maxFileStreamSize int, port string) {

	if maxMsgStreamSize <= 0 || maxFileStreamSize <= 0 {
		log.Fatal("max stream size must be greater than 0")
	}

	if maxConns <= 0 {
		log.Fatal("max connection must be greater than 0")
	}

	var conns = make(chan struct{}, maxConns)
	var cErrs = make(chan error, 100)
	var cResults = make(chan string)

	go func() {
		for e := range cErrs {
			log.Println(e)
		}
	}()

	go func() {
		for r := range cResults {
			log.Println("Received: ", r)
		}
	}()

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("error acepting conection: ", err)
			continue
		}

		conns <- struct{}{}
		go func() {
			defer func() { <-conns }()
			result, err := handleConnection(conn, maxMsgStreamSize, maxFileStreamSize)
			if err != nil {
				cErrs <- err
				return
			}
			cResults <- result
		}()
	}
}

func handleConnection(conn net.Conn, maxMsgStreamSize, maxFileStreamSize int) (string, error) {

	defer conn.Close()

	var headerSize [8]byte
	var headerType [4]byte

	var stremSize uint64 = 0
	var streamType uint32 = 0

	if _, err := io.ReadFull(conn, headerType[:]); err != nil {
		conn.Write([]byte("The peer encountered an issue while processing the stream.\n"))
		return "", err
	}
	if _, err := io.ReadFull(conn, headerSize[:]); err != nil {
		conn.Write([]byte("The peer encountered an issue while processing the stream.\n"))
		return "", err
	}
	streamType = binary.BigEndian.Uint32(headerType[:])

	if streamType != 1 && streamType != 2 {
		conn.Write([]byte("This peer does not recognize this type of stream.\n"))
		return "", errs.ErrUnknowStreamType
	}

	stremSize = binary.BigEndian.Uint64(headerSize[:])

	if stremSize == 0 {
		conn.Write([]byte("This peer does not support streams with zero size.\n"))
		return "", errs.ErrStreamSizeZero
	}

	if stremSize > uint64(maxMsgStreamSize) && streamType == 1 {
		conn.Write([]byte("This peer does not support such large streams for messages.\n"))
		return "", errs.ErrStreamToLarge
	}

	if stremSize > uint64(maxFileStreamSize) && streamType == 2 {
		conn.Write([]byte("This peer does not support such large streams for files.\n"))
		return "", errs.ErrStreamToLarge
	}

	result, err := readFromConnWithProgress(conn, stremSize, streamType, "test_name", "/home/dario61k/Downloads")
	if err != nil {
		conn.Write([]byte("The peer encountered an issue while processing the stream.\n"))
		return "", err
	}

	if _, err := conn.Write([]byte("ACK\n")); err != nil {
		return "", err
	}

	fmt.Println(result)

	return result, nil

}

func readFromConnWithProgress(conn net.Conn, stremSize uint64, streamType uint32, streamName, downloadPath string) (string, error) {

	const chunckSize = 32 * 1024
	var stream [chunckSize]byte

	typeTwo := streamType == 2

	var read uint64

	var f *os.File
	var m = make([]byte, 0, stremSize)

	var path string

	if typeTwo {
		path = filepath.Join(downloadPath, streamName)

		var err error
		f, err = os.Create(path)
		if err != nil {
			return "", err
		}
	}

	for read < stremSize {
		remaining := stremSize - read
		chunk := min(uint64(chunckSize), remaining)
		var n int
		var err error

		if typeTwo {
			n, err = conn.Read(stream[:chunk])

			if n > 0 {
				if _, err = f.Write(stream[:n]); err != nil {
					return "", err
				}
			}

			if err != nil {
				return "", err
			}

		} else {
			n, err = conn.Read(stream[:chunk])

			if n > 0 {
				m = append(m, stream[:n]...)
			}

			if err != nil {
				return "", err
			}

		}

		read += uint64(n)

		progress := int(read * 100 / stremSize)

		if typeTwo {
			fmt.Printf("\rProgress: %d%%", progress)
		}

	}

	if typeTwo {
		fmt.Println(path)
		return path, nil
	}

	return string(m), nil
}
