package main

import (
	"log"
	s "strings"
)

func HandleMessage(ws *WSConn, ch <-chan Message) {
	defer log.Println("loop closed somehow")
	channels := slurpFile("channels.txt")
	for {
		m := <-ch
		d := s.Split(m.Data, " ")
		ld := s.Split(s.ToLower(m.Data), " ")
		if m.Nick == "dbc__" || m.Nick == "tenseyi" {
			switch m.Command {
			case "MSG":
				if s.EqualFold(d[0], "!join") {
					if !inSlice(ld[1], channels) {
						channels = append(channels, ld[1])
						writeFile("channels.txt", channels)
						ws.Write("JOIN #" + ld[1])
						ws.Write("PRIVMSG #" + m.Channel + " :Logging " + ld[1])
					}
					ws.Write("PRIVMSG #" + m.Channel + " :Already logging " + ld[1])
					continue
				}
				if s.EqualFold(d[0], "!leave") {
					if inSlice(ld[1], channels) {
						channels = remove(ld[1], channels)
						writeFile("channels.txt", channels)
						ws.Write("PART #" + ld[1])
						ws.Write("PRIVMSG #" + m.Channel + " :Leaving " + ld[1])
					}
				}
			}
		}
	}
}
