package wire

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"strings"

	"drift/internal/fs"
)

func Send(address, req string) error {

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	defer conn.Close()

	var headerType [4]byte
	var headerInfoSize [8]byte
	var headerContentSize [8]byte
	
	var streamName string

	reqPrefix, reqPath, _ := strings.Cut(req, ":")

	content := []byte(req)
	
	binary.BigEndian.PutUint32(headerType[:], 1)
	
	typeTwo := reqPrefix == "@file" && reqPath != ""
	
	if typeTwo {

		binary.BigEndian.PutUint32(headerType[:], 2)
		file, err := fs.GetFileInfo(reqPath)
		if err != nil {
			return err
		}
		streamName = file.Name()

		binary.BigEndian.PutUint64(headerInfoSize[:], uint64(len([]byte(streamName))))

		f, err := fs.OpenFile(reqPath)
		if err != nil {
			return err
		}

		content = f
	}

	binary.BigEndian.PutUint64(headerContentSize[:], uint64(len(content)))

	pack := make([]byte, 0, len(headerType)+len(headerContentSize)+len(content))
	pack = append(pack, headerType[:]...)
	if typeTwo {
		pack = append(pack, headerInfoSize[:]...)
	}
	pack = append(pack, headerContentSize[:]...)
	if typeTwo {
		pack = append(pack, []byte(streamName)...)
	}
	pack = append(pack, content...)

	n, err := conn.Write(pack)
	
	if err != nil {
		log.Println("Write:", err)
		response, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Println("Read response:", err)
			return err
		}
		fmt.Print(response)
		return err
	}

	if n < len(pack) {
		return fmt.Errorf("error: incomplete transfer")
	}

	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		log.Println("Read response:", err)
		return err
	}

	fmt.Print(response)

	return nil

}