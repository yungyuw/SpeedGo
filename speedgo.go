package main

import (
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"strconv"
	"time"
)

func estimate(times uint) uint {
	return times*(3*54+8)*8 + 62
}

func ping(ws *websocket.Conn) {
	var what string
	err := websocket.Message.Receive(ws, &what)
	if err != nil {
		log.Printf("%s create an error at ping : %v\n", ws.Request().RemoteAddr, err)
		return
	}
	ctime := time.Now().UnixNano() / 1000000
	rtime, err := strconv.ParseInt(what, 10, 64)
	if err != nil {
		log.Printf("%s create an error when parse value : %v\n", ws.Request().RemoteAddr, err)
		return
	}

	websocket.Message.Send(ws, strconv.FormatInt(ctime, 10))

    log.Println("==============")
	log.Println(ws.Request().RemoteAddr, "%s query a ping/pong")
    log.Println("server timestamp :", ctime)
    log.Println("client timestamp :", what)
    log.Println("period :", ctime-rtime)
    log.Println("==============")

    err = websocket.Message.Receive(ws, &what)

	if err != nil {
		log.Printf("%s create an error when receive ping from client: %v\n", ws.Request().RemoteAddr, err)
		return
	}

    log.Printf("%s feedback a ping/pong result for %v ms\n", ws.Request().RemoteAddr, what)
}

func Handler(ws *websocket.Conn) {
	var counter uint64
	var times uint
	var what string
	var msg []byte
	err := websocket.Message.Receive(ws, &what)
	if err != nil {
		log.Println("%s create an error : %v", ws.Request().RemoteAddr, err)
	}
	start := time.Now()
	if what == "UPLOAD" {
		log.Printf("A client come for upload : %s\n", ws.Request().RemoteAddr)
		for {

			err := websocket.Message.Receive(ws, &msg)
			if err == nil {
				elapsed := time.Since(start)
				if elapsed.Seconds() < 1 {
					continue
				}
				counter += uint64(len(msg))
				times++
				if (len(msg) == 1 && msg[0] == 101) || elapsed.Seconds() > 11 {
					log.Println("--------From-------")
					log.Printf("Upload result for %s\n", ws.Request().RemoteAddr)
					log.Println(counter)
					period := elapsed.Seconds() - 1
					log.Printf("%vs\n", period)
					t := myoutput(float64(counter*8)/period + float64(estimate(times-1))/period)
					log.Println("--------End--------")
					websocket.Message.Send(ws, t)
					break
				}
			} else {
				log.Println("Error: ", err, " From ", ws.Request().RemoteAddr)
				return
			}
		}
	} else if what == "DOWNLOAD" {
		log.Printf("A client come for download : %s\n", ws.Request().RemoteAddr)
		var smsg string
		for i := 0; i < 10000; i++ {
			smsg += "D"
		}
		start = time.Now()
		counter = 0
		times = 0
		for {
			err := websocket.Message.Send(ws, smsg)
			if err == nil {
				elapsed := time.Since(start)
				if elapsed.Seconds() < 1 {
					continue
				}
				counter += uint64(len(smsg))
				times++
				if elapsed.Seconds() > 11 {
					log.Println("--------From-------")
					log.Printf("Download result for %s\n", ws.Request().RemoteAddr)
					log.Println(counter)
					period := elapsed.Seconds() - 1
					log.Printf("%vs\n", period)
					t := myoutput(float64(counter*8)/period + float64(estimate(times-1))/period)
					log.Println("--------End--------")
					websocket.Message.Send(ws, t)
					break
				}
			} else {
				log.Println("Error: ", err, " From ", ws.Request().RemoteAddr)
				return
			}
		}
	} else {
		log.Printf("Other from %s\n", ws.Request().RemoteAddr)
	}
}

func myoutput(num float64) string {
	var times int
	for num > 1024 && times < 4 {
		num /= 1024
		times++
	}
	log.Printf("%.3f\n", num)
	var t string
	switch times {
	case 1:
		t = "K"
	case 2:
		t = "M"
	case 3:
		t = "G"
	case 4:
		t = "T"
	}
	log.Println(t)
	return strconv.FormatFloat(num, 'f', 6, 64) + t
}

func main() {
	log.SetFlags(log.Lshortfile)
	log.Println("Server Start")
	http.Handle("/ping", websocket.Handler(ping))
	http.Handle("/echo", websocket.Handler(Handler))
	log.Fatal(http.ListenAndServe(":8090", nil))
}
