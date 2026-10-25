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
	"drift/internals/fs"
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

	var headerType [4]byte
	var headerInfoSize [8]byte
	var headerContentSize [8]byte

	var streamType uint32 = 0
	var streamInfoSize uint64 = 0
	var streamContentSize uint64 = 0

	if _, err := io.ReadFull(conn, headerType[:]); err != nil {
		conn.Write([]byte(errs.ErrDefault.Error()))
		return "", err
	}

	streamType = binary.BigEndian.Uint32(headerType[:])

	if streamType != 1 && streamType != 2 {
		conn.Write([]byte(errs.ErrUnknowStreamType.Error()))
		return "", errs.ErrUnknowStreamType
	}

	if streamType == 2 {
		if _, err := io.ReadFull(conn, headerInfoSize[:]); err != nil {
			conn.Write([]byte(errs.ErrDefault.Error()))
			return "", err
		}

		streamInfoSize = binary.BigEndian.Uint64(headerInfoSize[:])

		if streamInfoSize > 100 {
			conn.Write([]byte(errs.ErrStreamInfoToLarge.Error()))
			return "", errs.ErrStreamInfoToLarge
		}
	}

	if _, err := io.ReadFull(conn, headerContentSize[:]); err != nil {
		conn.Write([]byte(errs.ErrDefault.Error()))
		return "", err
	}

	streamContentSize = binary.BigEndian.Uint64(headerContentSize[:])

	if streamContentSize == 0 {
		conn.Write([]byte(errs.ErrStreamSizeZero.Error()))
		return "", errs.ErrStreamSizeZero
	}

	if streamContentSize > uint64(maxMsgStreamSize) && streamType == 1 {
		conn.Write([]byte(errs.ErrStreamToLarge.Error()))
		return "", errs.ErrStreamToLarge
	}

	if streamContentSize > uint64(maxFileStreamSize) && streamType == 2 {
		conn.Write([]byte(errs.ErrStreamToLarge.Error()))
		return "", errs.ErrStreamToLarge
	}

	result, err := readFromConnWithProgress(conn, streamContentSize, streamInfoSize, streamType, "/home/dario61k/Downloads")
	if err != nil {
		conn.Write([]byte(errs.ErrDefault.Error()))
		return "", err
	}

	if _, err := conn.Write([]byte("ACK\n")); err != nil {
		return "", err
	}

	return result, nil

}

func readFromConnWithProgress(conn net.Conn, streamContentSize, streamInfoSize uint64, streamType uint32, downloadPath string) (string, error) {

	const chunckSize = 32 * 1024
	var stream [chunckSize]byte

	typeTwo := streamType == 2

	var read uint64

	var f *os.File
	var m = make([]byte, 0, streamContentSize)

	var path string
	var streamName string

	if typeTwo {

		var err error
		info := make([]byte, streamInfoSize)

		_, err = io.ReadFull(conn, info[:streamInfoSize])
		if err != nil {
			return "", err
		}

		streamName = string(info)
		if !fs.IsValidFilename(streamName) {
			return "", errs.ErrInvalidStreamInfo
		}

		path = filepath.Join(downloadPath, streamName)

		
		f, err = os.Create(path)
		if err != nil {
			return "", err
		}

		defer f.Close()
	}

	for read < streamContentSize {
		remaining := streamContentSize - read
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

		progress := int(read * 100 / streamContentSize)

		if typeTwo {
			fmt.Printf("\r%s : %d%%", streamName, progress)
		}
	}

	if typeTwo {
		return path, nil
	}

	return string(m), nil
}