package main

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/gdamore/tcell"

	"github.com/redbo/mudengine/headlesstcell"
	"golang.org/x/crypto/ssh"
)

type windowResizer interface {
	WindowResize(w, h int)
}

func main() {
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)

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

		var term tcell.Screen

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
					term, err = headlesstcell.NewScreen(channel,
						termName, int(cols), int(rows))
					if err := term.Init(); err != nil {
						req.Reply(false, nil)
					} else {
						req.Reply(true, nil)
						go func() {
							defer channel.Close()
							run(term)
						}()
					}
				case "window-change":
					cols := binary.BigEndian.Uint32(req.Payload[0:4])
					rows := binary.BigEndian.Uint32(req.Payload[4:8])
					if wr, ok := term.(windowResizer); ok {
						wr.WindowResize(int(cols), int(rows))
					}
				}
			}
		}(requests)
	}
}

func run(s tcell.Screen) {
	s.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorWhite))
	s.Clear()

	quit := make(chan struct{})
	go func() {
		for {
			ev := s.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyEscape, tcell.KeyEnter:
					close(quit)
					return
				case tcell.KeyCtrlL:
					s.Sync()
				}
			case *tcell.EventResize:
				s.Sync()
			}
		}
	}()

	cnt := 0
	dur := time.Duration(0)
loop:
	for {
		select {
		case <-quit:
			break loop
		case <-time.After(time.Millisecond * 50):
		}
		start := time.Now()
		makebox(s)
		cnt++
		dur += time.Now().Sub(start)
	}

	s.Fini()
	fmt.Printf("Finished %d boxes in %s\n", cnt, dur)
	fmt.Printf("Average is %0.3f ms / box\n", (float64(dur)/float64(cnt))/1000000.0)
}

func makebox(s tcell.Screen) {
	w, h := s.Size()

	if w == 0 || h == 0 {
		return
	}

	glyphs := []rune{'@', '#', '&', '*', '%', 'Z', 'A', ' '}

	lx := rand.Int() % w
	ly := rand.Int() % h
	lw := rand.Int() % (w - lx)
	lh := rand.Int() % (h - ly)
	st := tcell.StyleDefault
	gl := ' '
	if s.Colors() > 256 {
		rgb := tcell.NewHexColor(int32(rand.Int() & 0xffffff))
		st = st.Background(rgb)
	} else if s.Colors() > 1 {
		st = st.Background(tcell.Color(rand.Int() % s.Colors()))
	} else {
		st = st.Reverse(rand.Int()%2 == 0)
	}
	gl = glyphs[rand.Int()%len(glyphs)]

	for row := 0; row < lh; row++ {
		for col := 0; col < lw; col++ {
			s.SetCell(lx+col, ly+row, st, gl)
		}
	}
	s.Show()
}
