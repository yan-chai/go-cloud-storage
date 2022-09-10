package main

import (
	"cloudStore/message"
	"cloudStore/utils"
	"bufio"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
)

/*func handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		decoder := gob.NewDecoder(conn)
		msg:= &message.Message{}
		decoder.Decode(msg)
		fmt.Println(msg)

		file,_ = os.OpenFile("newfile.txt", os.O_CREATE | os.O_TRUNC | os.O_RDWR, 0666)
		io.CopyN(file, conn, 1024)
		
	}
}*/

type MessageHandler func(*message.Message, net.Conn, io.Reader)

var handlers = map[message.MessageType]MessageHandler{
	"put":          handlePutReq,
	"get":          handleGetRequest,
	"delete":       handleDeleteRequest,
	"search":       handleSearchRequest,
}

func handleDeleteBackup(msg *message.Message, conn net.Conn, reader io.Reader) {
	defer conn.Close()
	if !fileExist(msg.Name) {
		reply := message.Response(false, "File does not exist", 0, nil)
		reply.Reply(conn)
	} else {
		err := os.Remove(msg.Name)
		utils.CheckError(err)
		err = os.Remove(msg.Name + ".checkSum")
		utils.CheckError(err)
		reply := message.Response(true, "File successfully deleted on the sever.", 0, nil)
		reply.Reply(conn)
	}
}

func handleBackup(msg *message.Message, conn net.Conn, reader io.Reader) {
	defer conn.Close()
	if fileExist(msg.Name) {
		reply := message.Response(false, "Duplicate file name exist, please use a different file name", 0, nil)
		reply.Reply(conn)
	} else {
		destFile, err := os.OpenFile(msg.Name, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
		utils.CheckError(err)
		io.CopyN(destFile, reader, msg.Head.Size)
		destFile.Close()

		checkSumFile, err := os.Create(msg.Name + ".checkSum")
		utils.CheckError(err)

		_, err = checkSumFile.WriteString(msg.Md5)
		utils.CheckError(err)
		checkSumFile.Close()

		checkSum, _ := utils.ComputeMd5(msg.Name)
		fmt.Println(checkSum)

		reply := message.Response(true, "File successfully added to the sever", 0, nil)
		reply.Reply(conn)
	}
}

func handleSearchRequest(msg *message.Message, conn net.Conn, reader io.Reader) {
	root := "./"
	var files []string
	defer conn.Close()
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if len(msg.Name) == 0 && !strings.Contains(info.Name(), "checkSum") {
			files = append(files, info.Name())
		} else if strings.Contains(info.Name(), msg.Name) && !strings.Contains(info.Name(), "checkSum") {
			files = append(files, info.Name())
		}

		return nil
	})
	utils.CheckError(err)

	if len(files) == 0 {
		reply := message.Response(false, "File does not exist", 0, nil)
		reply.Reply(conn)
	} else {
		reply := message.Response(true, "Matching Files Found", 0, files)
		reply.Reply(conn)
	}
}

func handleDeleteRequest(msg *message.Message, conn net.Conn, reader io.Reader) {
	defer conn.Close()
	if !fileExist(msg.Name) {
		reply := message.Response(false, "File does not exist", 0, nil)
		reply.Reply(conn)
	} else {
		deleteBackUpResponse := deleteBakcup(msg.Name, msg.BackUpAddr)
		processBackUpResponse(deleteBackUpResponse)

		if deleteBackUpResponse.Success {
			err := os.Remove(msg.Name)
			utils.CheckError(err)
			err = os.Remove(msg.Name + ".checkSum")
			utils.CheckError(err)
			reply := message.Response(true, "File successfully deleted on the sever.", 0, nil)
			reply.Reply(conn)
		} else {
			reply := message.Response(false, "BackUp Server fails to delete, file is not deleted.", 0, nil)
			reply.Reply(conn)
		}
	}
}

func handleGetRequest(msg *message.Message, conn net.Conn, reader io.Reader) {
	defer conn.Close()
	fileWriter := bufio.NewWriter(conn)
	if !fileExist(msg.Name) {
		reply := message.Response(false, "File does not exist", 0, nil)
		reply.Reply(conn)
	} else {
		if isFileOK(msg.Name) {
			sendFileToClient(msg.Name, conn, fileWriter)
		} else {
			replyFromBackUp := retrieveFromBackup(msg.Name, msg.BackUpAddr)
			processBackUpResponse(replyFromBackUp)
			if replyFromBackUp.Success {
				sendFileToClient(msg.Name, conn, fileWriter)
			} else {
				reply := message.Response(false, "BackUp Server fails, file is not stored.", 0, nil)
				reply.Reply(conn)
			}
		}
	}
}

func sendFileToClient(fileName string, conn net.Conn, writer *bufio.Writer) {
	fileInofo, err := os.Stat(fileName)
	utils.CheckError(err)
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0666)
	utils.CheckError(err)
	reply := message.Response(true, "File successfully retrieved", fileInofo.Size(), nil)
	reply.Reply(conn)
	io.Copy(writer, file)
	writer.Flush()
}

func handlePutReq(msg *message.Message, conn net.Conn, reader io.Reader) {
	defer conn.Close()
	if fileExist(msg.Name) {
		reply := message.Response(false, "Duplicate file name exist, please use a different file name", 0, nil)
		reply.Reply(conn)
	} else {
		destFile, err := os.OpenFile(msg.Name, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
		utils.CheckError(err)
		io.CopyN(destFile, reader, msg.Head.Size)
		destFile.Close()

		checkSumFile, err := os.Create(msg.Name + ".checkSum")
		utils.CheckError(err)

		_, err = checkSumFile.WriteString(msg.Md5)
		utils.CheckError(err)
		checkSumFile.Close()

		checkSum, _ := utils.ComputeMd5(msg.Name)
		fmt.Println(checkSum)

		replyFromBackUp := backup(msg.Name, msg.BackUpAddr)
		processBackUpResponse(replyFromBackUp)

		if replyFromBackUp.Success {
			reply := message.Response(true, "File successfully added to the sever.", 0, nil)
			reply.Reply(conn)
		} else {
			err := os.Remove(msg.Name)
			utils.CheckError(err)
			reply := message.Response(false, "BackUp Server fails, file is not stored.", 0, nil)
			reply.Reply(conn)
		}

	}
}

func processBackUpResponse(response *message.MessageReply) {
	if response.Success {
		fmt.Println("Operation Success! " + response.Response)
	} else {
		fmt.Println("Operation Fail! " + response.Response)
	}
}

func backup(fileName string, serverAddr string) *message.MessageReply {
	srcFile, err := os.OpenFile(fileName, os.O_RDONLY, 0666)
	utils.CheckError(err)
	conn, err := net.Dial("tcp", serverAddr)
	utils.CheckError(err)

	defer conn.Close()

	fileInofo, err := os.Stat(fileName)
	utils.CheckError(err)

	checkSum, err := utils.ComputeMd5(fileName)
	utils.CheckError(err)

	msg := message.New("backup", fileInofo.Size(), fileName, checkSum, "")
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
	return reply
}
func retrieveFromBackup(fileName string, serverAddr string) *message.MessageReply {
	conn, err := net.Dial("tcp", serverAddr)
	utils.CheckError(err)

	defer conn.Close()

	msg := message.New("get", 0, fileName, "", serverAddr)
	msg.Send(conn)

	reader := bufio.NewReader(conn)
	decoder := gob.NewDecoder(reader)

	reply := &message.MessageReply{}
	decoder.Decode(reply)
	processBackUpResponse(reply)
	if reply.Success {
		err := os.Remove(fileName)
		utils.CheckError(err)
		err = os.Remove(fileName + ".checkSum")
		utils.CheckError(err)
		destFile, err := os.Create(fileName)
		utils.CheckError(err)
		io.CopyN(destFile, reader, reply.FileSize)

		checkSumFile, err := os.Create(fileName + ".checkSum")
		utils.CheckError(err)
		checkSum, _ := utils.ComputeMd5(fileName)
		fmt.Println(checkSum)
		_, err = checkSumFile.WriteString(checkSum)
		utils.CheckError(err)
		checkSumFile.Close()
	}
	return reply
}

func deleteBakcup(fileName string, serverAddr string) *message.MessageReply {
	conn, err := net.Dial("tcp", serverAddr)
	utils.CheckError(err)

	defer conn.Close()
	msg := message.New("deleteBackup", 0, fileName, "", "")
	msg.Send(conn)
	reply := &message.MessageReply{}
	decoder := gob.NewDecoder(conn)
	decoder.Decode(reply)
	return reply
}

func fileExist(fileName string) bool {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		fmt.Println(err)
		return false
	}
	return true
}

func isFileOK(fileName string) bool {
	data, err := ioutil.ReadFile(fileName + ".checkSum")
	utils.CheckError(err)
	checkSum, err := utils.ComputeMd5(fileName)
	utils.CheckError(err)
	strCheckSum := string(data)
	return strCheckSum == checkSum
}

func main() {
	listener, err := net.Listen("tcp", ":9999")
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		if conn, err := listener.Accept(); err == nil {
			reader := bufio.NewReader(conn)
			decoder := gob.NewDecoder(reader)
			msg := &message.Message{}
			decoder.Decode(msg)
			request := handlers[msg.Head.Type]
			go request(msg, conn, reader)
		}
	}
}