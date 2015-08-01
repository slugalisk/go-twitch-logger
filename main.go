package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	s "strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WSConn struct {
	sync.Mutex
	channel string
	conn    *websocket.Conn
	dialer  websocket.Dialer
	heads   http.Header
}

type Message struct {
	Channel   string
	Nick      string
	Data      string
	Timestamp time.Time
	Command   string
}

var (
	chLog  = make(chan Message, 10)
	url    = "ws://irc.twitch.tv:80/ws"
	origin = []string{"http://irc.twitch.tv"}
)

func (w *WSConn) Reconnect() {
	if w.conn != nil {
		w.Lock()
		_ = w.conn.Close()
		w.Unlock()
	}

	var err error
	w.conn, _, err = w.dialer.Dial(url, w.heads)
	if err != nil {
		logAndDelay(err, w.channel)
		w.Reconnect()
	}

	w.Write("PASS oauth:")
	w.Write("NICK bla")

	channels := slurpFile("channels.txt")
	log.Println("Connecting to", len(channels), "channels")
	go func() {
		for _, ch := range channels {
			w.Write("JOIN #" + s.ToLower(ch))
			log.Println("connected to", ch)
			time.Sleep(500 * time.Millisecond)
		}
	}()
}

func (w *WSConn) Read() []Message {
	w.conn.SetReadDeadline(time.Now().Add(120 * time.Minute))
	_, msg, err := w.conn.ReadMessage()
	if err != nil {
		logAndDelay(err, w.channel)
		w.Reconnect()

	}
	if s.Index(string(msg), "PING") == 0 {
		w.Write(s.Replace(string(msg), "PING", "PONG", -1))
	}
	return parse(msg)
}

func (w *WSConn) Write(msg string) {
	w.Lock()
	err := w.conn.WriteMessage(1, []byte(msg+"\r\n"))
	w.Unlock()
	if err != nil {
		logAndDelay(err, w.channel)
		w.Reconnect()
	}
}

//ACTION jjk
// :lorray!lorray@lorray.tmi.twitch.tv PRIVMSG #reckful :@Reckful DO THE TUTORIAL YOU FUCKIN PLEB
func parse(msg []byte) (m []Message) {
	if s.Contains(string(msg), "PRIVMSG #") {
		re, _ := regexp.Compile(`:(.+)!.+tmi\.twitch\.tv PRIVMSG #([a-z0-9_-]+) :(.+)`)
		l := re.FindAllStringSubmatch(string(msg), -1)
		for _, v := range l {
			var mm Message
			mm.Command = "MSG"
			mm.Nick = s.TrimSpace(v[1])
			mm.Channel = s.TrimSpace(v[2])
			mm.Data = s.Replace(s.Replace(s.TrimSpace(v[3]), "ACTION", "/me", -1), "", "", -1)
			mm.Timestamp = time.Now()
			m = append(m, mm)
		}
		return m
	}
	return m
}

func LogErr(err error) {
	if err != nil {
		log.Printf("ERROR: %s\n", err)
	}
}

func logAndDelay(err error, name string) {
	log.Printf("%s > Connection failed ERROR: %s\n", name, err)
	log.Println("Reconnecting in 1 Seconds...")
	time.Sleep(1 * time.Second)
}

func slurpFile(fn string) []string {
	d, err := ioutil.ReadFile(fn)
	if err != nil {
		LogErr(err)
		return []string{}
	}
	dl := s.Split(string(d), ",")
	var dn []string
	for _, v := range dl {
		if v != "" {
			dn = append(dn, v)
		}
	}
	return dn
}

func inSlice(n string, l []string) bool {
	for _, v := range l {
		if s.EqualFold(v, n) {
			return true
		}
	}
	return false
}

func writeFile(fn string, s []string) {
	var d string

	for _, v := range s {
		if v != "" {
			d += v + ","
		}
	}

	f, err := os.Create(fn)
	if err != nil {
		LogErr(err)
		return
	}
	defer f.Close()
	f.WriteString(d)
}

func remove(n string, sl []string) []string {
	for i, data := range sl {
		if s.EqualFold(n, data) {
			sl = append(sl[:i], sl[i+1:]...)
			return sl
		}
	}
	return sl
}

func main() {
	// log.SetFlags(log.LstdFlags | log.Lshortfile)
	chLog := make(chan Message, 100)
	chCMD := make(chan Message, 100)

	ws := &WSConn{
		dialer: websocket.Dialer{HandshakeTimeout: 20 * time.Second},
	}

	ws.Reconnect()

	go WriteFile(chLog)
	go HandleMessage(ws, chCMD)
	for {
		msg := ws.Read()
		for _, v := range msg {
			log.Println("[CHAT]", v.Channel, ">", v.Nick+":", v.Data)
			chLog <- v
			chCMD <- v
		}
	}
}
