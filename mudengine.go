package main

// TERMDASH

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

func main() {
	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			// Should use constant-time compare (or better, salt+hash) in
			// a production setting.
			fmt.Println(c.User(), string(pass))
			return nil, nil
			// return nil, fmt.Errorf("password rejected for %q", c.User())
		},
	}

	privateBytes, err := ioutil.ReadFile("id_rsa")
	if err != nil {
		log.Fatal("Failed to load private key: ", err)
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatal("Failed to parse private key: ", err)
	}

	config.AddHostKey(private)

	listener, err := net.Listen("tcp", "0.0.0.0:2022")
	if err != nil {
		log.Fatal("failed to listen for connection: ", err)
	}
	nConn, err := listener.Accept()
	if err != nil {
		log.Fatal("failed to accept incoming connection: ", err)
	}

	_, chans, reqs, err := ssh.NewServerConn(nConn, config)
	if err != nil {
		log.Fatal("failed to handshake: ", err)
	}

	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Fatalf("Could not accept channel: %v", err)
		}

		term := newTerminal(channel)

		go func(in <-chan *ssh.Request) {
			for req := range in {
				fmt.Println(req.Type)
				switch req.Type {
				case "shell":
					req.Reply(true, nil)
				case "pty-req":
					fmt.Println(string(req.Payload))
					termLen := req.Payload[3]
					termName := string(req.Payload[4 : termLen+4])
					cols := binary.BigEndian.Uint32(req.Payload[termLen+4 : termLen+8])
					rows := binary.BigEndian.Uint32(req.Payload[termLen+8 : termLen+12])
					fmt.Println(term, cols, rows)
					term.Init(int(cols), int(rows), termName)
					req.Reply(true, nil)
				case "window-change":
					cols := binary.BigEndian.Uint32(req.Payload[0:4])
					rows := binary.BigEndian.Uint32(req.Payload[4:8])
					term.SetSize(int(cols), int(rows))
					fmt.Println("Window change", term, cols, rows)
				}
			}
		}(requests)

		go func() {
			defer channel.Close()
			for {
				err := term.ProcessInput()
				if err != nil {
					break
				}
			}
		}()
	}
}
