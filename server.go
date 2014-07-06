package main

import (
	"fmt"
	"net/http"
	"code.google.com/p/go.net/websocket"
	"time"
	"encoding/json"
	"math/rand"
	"flag"
	"github.com/dennisfrancis/ntee"
)


var port *int = flag.Int("port", 1234, "Port to listen")
var root *string = flag.String("root", ".", "root to serve")

type RANDREQ struct {
	Len int
	Retchan chan string
}

var randreq_chan = make(chan RANDREQ)

var bindata = []byte{}  // This is the byte array to be sent through websocket

var in chan<- bool
var add_chan chan<- chan bool
var del_chan chan<- chan bool

func main() {

	flag.Parse()
	nt := ntee.New()
	in = nt.GetInputChan()
	add_chan = nt.GetOutputAdder()
	del_chan = nt.GetOutputDeleter()
	nt.Run()

	go RandString()
	go gendata()

	http.Handle("/", http.FileServer(http.Dir(*root)))
	http.Handle("/tabledata", websocket.Handler(dataFeeder))
	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
	return

}


func dataFeeder(ws *websocket.Conn) {
	fmt.Println("Received a ws connection")
	trigger_channel := make(chan bool)
	add_chan <- trigger_channel
	fmt.Println("Added trigger_channel")

	for _ = range trigger_channel {
		fmt.Println("data-generator updated bindata.")
		err := websocket.Message.Send(ws, bindata)
		fmt.Println("Size of bindata =", len(bindata))
		if err != nil {  // Client closed conn ?
			fmt.Println("Send returned err = ", err)
			del_chan <- trigger_channel
			break
		}
	}
	return
}


func gendata() {  // This function generates/updates data and puts in bindata variable.
	count := 1
	var data = make(map[string]interface{})

	var cols = make([]string, 10)
	var rows = make([][]string, 4000)

	for ii := 0; ii<10; ii++ {
		cols[ii] = fmt.Sprintf("col-%d", ii)
	}

	for ii := 0; ii<4000; ii++ {
		rows[ii] = make([]string, 10)
	}

	data["cols"] = cols
	retchan := make(chan string)

	for {

		for ii := 0; ii<4000; ii++ {

			for jj := 0; jj<10; jj++ {
				randreq_chan <- RANDREQ{Len : 10, Retchan : retchan}
				rows[ii][jj] = <-retchan
			}

		}

		data["rows"] = rows
		data["count"] = fmt.Sprintf("%d", count)

		// fill bindata
		bindata, _ = json.Marshal(data)

		// Sent msg to ntee module to signal the update
		in <- true

		time.Sleep(5*time.Second)
		count += 1
	}

}

func RandString() {
	rndbuf := make([]byte, 1024)
	var alph string = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	for req := range randreq_chan {
		for i:=0; i<req.Len; i++ {
			rndbuf[i] = alph[rand.Intn(36)]
		}
		req.Retchan <- string(rndbuf[0:req.Len])
	}
}
