package main

import (
	"github.com/gopherjs/jquery"
	"github.com/gopherjs/gopherjs/js"
	"encoding/json"
	"time"
	"honnef.co/go/js/console"
)

var jQuery = jquery.NewJQuery
var document js.Object
var WebSocket  =  js.Global.Get("WebSocket")
var connected bool = true
var connection js.Object

func main() {

	jQuery("document").Ready(startup)

	document =  js.Global.Get("window").Get("document")
	
}


func startup() {
	jQuery("#tab-container").Underlying().Call("easytabs");
	jQuery("#tab-container").Underlying().Call("bind", "easytabs:ajax:complete", create_table)
	
}

func create_table() {

	connection  =  WebSocket.New("ws://localhost:1234/tabledata")
	connection.Set("binaryType", "arraybuffer")

	first := true
	var tabledata []map[string]string
	var columns []map[string]string
	var options = map[string]interface{}{
		"enableCellNavigation" : true,
		"enableColumnReorder"  : false,
		"fullWidthRows"        : true,
		"defaultColumnWidth"   : 150,
	}

	var sg js.Object

	ws_onopen := func() {
		console.Log("WS opened")
	}

	ws_onerror := func(err js.Object) {
		console.Log("WS error = ",  err);
	}

	ws_onmessage := func(e js.Object) {

		t0 := time.Now()
		data := map[string]interface{}{}
		buf  := js.Global.Get("Uint8Array").New(e.Get("data")).Interface().([]byte)
		console.Log("WS received byte stream of length = %d", len(buf))
		err := json.Unmarshal(buf, &data)

		if err != nil {
			console.Log("unmarshal error : ", err.Error())
			return
		}
		rows  := data["rows"].([]interface{})
		cols  := data["cols"].([]interface{})
		count := data["count"].(string)
		console.Log("count = %s", count)

		if first == true {
			columns = make([]map[string]string, len(cols))
			tabledata = make([]map[string]string, len(rows))			
		}
		for ii, tmp := range cols {
			name := tmp.(string)
			columns[ii] = map[string]string{"id" : name, "name" : name, "field" : name}
		}
	

		for ii:=0; ii<len(rows); ii++ {
			tabledata[ii] = map[string]string{}
			for jj:=0; jj<len(cols); jj++ {
				tabledata[ii][columns[jj]["field"]] = rows[ii].([]interface{})[jj].(string)
			}
		}
		t1 := time.Now()
		console.Log("pre-rendering time = %.3f ms", t1.Sub(t0).Seconds()/1000.0)
		if first == true {
			sg = js.Global.Get("Slick").Get("Grid").New("#alerts-table", tabledata, columns, options)
			
		} else {
			sg.Call("setData", tabledata, false)
			sg.Call("render")
		}
		t2 := time.Now()
		console.Log("rendering time = %.3f ms", t2.Sub(t1).Seconds()/1000.0)
		first = false
	}

	// Set the callbacks
	connection.Set("onopen", ws_onopen)
	connection.Set("onerror", ws_onerror)
	connection.Set("onmessage", ws_onmessage)
	jQuery("#btcontrol").On(jquery.CLICK, button_clicked)
	connected = true

}


func button_clicked(e jquery.Event) {
	button := jQuery("#btcontrol")
	if connected == true {
		console.Log("Clicked Close Connection")
		button.SetProp("disabled", true)
		connection.Call("close", 3000)
		console.Log("connection.readyState =", connection.Get("readyState"))
		connected = false
		button.SetText("Open Connection")
		button.SetProp("disabled", false)
	} else {
		console.Log("Clicked Open Connection")
		button.SetProp("disabled", true)
		create_table()
		connected = true
		button.SetText("Close Connection")
		button.SetProp("disabled", false)
	}
}
