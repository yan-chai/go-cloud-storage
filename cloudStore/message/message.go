package message

import (
	"fmt"
	"net"
	"encoding/gob"
)

type MessageType string

/*const (
	StorageRequest MessageType = iota
	RetrievalRequest
	SearchRequest
)*/

type MessageHead struct {
	Size int64
	Type MessageType
}

type Message struct {
	Head MessageHead
	BackUpAddr string
	Name string
	Md5 string
}

type MessageReply struct {
	Success bool
	Response string
	FileSize int64
	Files []string
}

func New(ty MessageType, size int64, name string, md5 string, backUpAddr string) *Message {
	head := MessageHead{size, ty}
	msg := Message{head, backUpAddr, name, md5}
	return &msg
}

func Response(issuccess bool, response string, fileSize int64, files []string) *MessageReply {
	msg := MessageReply{issuccess, response, fileSize, files}
	return &msg
}

func (m *Message) Print(){
	fmt.Println(m)
}

func (m *Message) Send(conn net.Conn) error {
	encoder := gob.NewEncoder(conn)
	err := encoder.Encode(m)
	return err
}

func (r *MessageReply) Reply(conn net.Conn) error {
	encoder := gob.NewEncoder(conn)
	err := encoder.Encode(r)
	return err
}