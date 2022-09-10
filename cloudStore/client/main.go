package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"cloudStore/message"
	"cloudStore/utils"
	"encoding/gob"
)

type requestHandler func(string, string, string)

func search(fileName string, serverAddr string, backupAddr string) {
	conn, err := net.Dial("tcp", serverAddr)
	utils.CheckError(err)

	defer conn.Close()
	msg := message.New("search", 0, fileName, "", backupAddr)
	msg.Send(conn)
	reply := &message.MessageReply{}
	decoder := gob.NewDecoder(conn)
	decoder.Decode(reply)
	processResponse(reply)

	if reply.Success {
		for _, file := range reply.Files {
			fmt.Println(file)
		}
	}
}

func delete(fileName string, serverAddr string, backupAddr string) {
	conn, err := net.Dial("tcp", serverAddr)
	utils.CheckError(err)

	defer conn.Close()
	msg := message.New("delete", 0, fileName, "", backupAddr)
	msg.Send(conn)
	reply := &message.MessageReply{}
	decoder := gob.NewDecoder(conn)
	decoder.Decode(reply)
	processResponse(reply)
}

func retrieve(fileName string, serverAddr string, backupAddr string) {
	conn, err := net.Dial("tcp", serverAddr)
	utils.CheckError(err)

	defer conn.Close()

	msg := message.New("get", 0, fileName, "", backupAddr)
	msg.Send(conn)

	reader := bufio.NewReader(conn)
	decoder := gob.NewDecoder(reader)

	reply := &message.MessageReply{}
	decoder.Decode(reply)
	processResponse(reply)
	if reply.Success {
		destFile, err := os.Create(fileName)
		utils.CheckError(err)
		io.CopyN(destFile, reader, reply.FileSize)
	}
}

func upload(fileName string, serverAddr string, backupAddr string) {
	srcFile, err := os.OpenFile(fileName, os.O_RDONLY, 0666)
	utils.CheckError(err)
	conn, err := net.Dial("tcp", serverAddr)
	utils.CheckError(err)

	defer conn.Close()

	fileInofo, err := os.Stat(fileName)
	utils.CheckError(err)

	checkSum, err := utils.ComputeMd5(fileName)
	utils.CheckError(err)

	msg := message.New("put", fileInofo.Size(), fileName, checkSum, backupAddr)
	msg.Send(conn)

	fileWriter := bufio.NewWriter(conn)
	encoder := gob.NewEncoder(fileWriter)

	_, err = io.Copy(fileWriter, srcFile)
	utils.CheckError(err)

	encoder.Encode(fileWriter)
	fileWriter.Flush()

	reply := &message.MessageReply{}
	decoder := gob.NewDecoder(conn)
	decoder.Decode(reply)
	processResponse(reply)

}

func processResponse(response *message.MessageReply) {
	if response.Success {
		fmt.Println("Operation Success! " + response.Response)
	} else {
		fmt.Println("Operation Fail! " + response.Response)
	}
}

var handlers = map[string]requestHandler{
	"put":    upload,
	"get":    retrieve,
	"delete": delete,
	"search": search,
}
func main() {

	if len(os.Args) != 4 {
		fmt.Printf("not enough arguments")
		os.Exit(0)
    }

	action := os.Args[1]
	var srcFile string
	if action == "search" && len(os.Args) == 2 {
		srcFile = ""
		fmt.Print("test")
	} else {
		srcFile = os.Args[2]
	}

	serverAddr := os.Args[3]
	backUpAddr := ""
	if serverAddr == "localhost:9998"{
		backUpAddr = "localhost:9999"
	} else{
		backUpAddr = "localhost:9998"
	}
	request := handlers[action]
	request(srcFile, serverAddr, backUpAddr)
    
	
}

