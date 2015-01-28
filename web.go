// web.go - web interface for beep
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
)

type Web struct {
	req  *http.Request
	res  http.ResponseWriter
	tmpl *template.Template
}

// Starts beep web server
func startWebServer() {
	var err error
	ip := "localhost" // serve locally by default
	port := 4444

	for i, arg := range os.Args {
		if i == 0 || strings.HasPrefix(arg, "-") {
			continue
		}
		parts := strings.Split(arg, ":")
		if len(parts) != 2 {
			continue
		}
		if len(parts[0]) > 0 {
			ip = parts[0]
		}
		if len(parts[1]) > 0 {
			port, err = strconv.Atoi(parts[1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid port number: %v", parts[1])
				os.Exit(1)
			}
		}
		break
	}
	address := fmt.Sprintf("%s:%d", ip, port)
	fmt.Printf("Listening on http://%s/\n", address)

	web := NewWeb()
	music.quietMode = true
	err = http.ListenAndServe(address, web)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to start web server:", err)
		os.Exit(1)
	}
}

func NewWeb() *Web {
	w := &Web{
		tmpl: template.Must(template.New("tmpl").Parse(webTemplates)),
	}
	return w
}

func (w *Web) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	w.req = req
	w.res = res
	w.serve()
}

// Serve pages
func (w *Web) serve() {
	defer func() {
		if obj := recover(); obj != nil {
			format := "\" ><pre>\nError: %s\nStack:%s\n</pre>"
			fmt.Fprintf(w.res, format, obj, debug.Stack())
		}
	}()
	path := w.req.URL.Path
	if file, found := webFileMap[path]; found {
		if strings.HasSuffix(path, ".css") {
			w.res.Header().Add("Content-Type", "text/css")
		}
		if strings.HasSuffix(path, ".js") {
			w.res.Header().Add("Content-Type", "text/javascript")
		}
		fmt.Fprint(w.res, file)
		return
	}
	defer w.req.Body.Close()
	switch path {
	case "/":
		w.serveHome()
	case "/play":
		w.servePlay()
	case "/stop":
		w.serveStop()
	case "/search":
		w.serveSearch()
	case "/loadSheet":
		w.serveLoadSheet()
	case "/saveSheet":
		w.serveSaveSheet()
	case "/voices":
		w.serveVoices()
	case "/downloadVoice":
		w.serveDownloadVoice()
	default:
		w.execTemplate("header", nil)
		w.execTemplate("pageNotFound", w.req.URL.Path)
	}
}

// Serves home page
func (w *Web) serveHome() {
	type homePage struct {
		Demo string
	}
	data := &homePage{
		Demo: demoMusic,
	}
	w.execTemplate("header", nil)
	w.execTemplate("/", data)
}

// Playback
func (w *Web) servePlay() {
	if music.playing {
		return
	}
	type playRequest struct {
		Notation string
	}
	request := &playRequest{}
	w.jsonRequest(request)
	initSoundDevice()
	notation := bytes.NewBuffer([]byte(request.Notation))
	reader := bufio.NewReader(notation)
	go playMusicNotes(reader, 100)
	<-music.played
}

// Stops playback
func (w *Web) serveStop() {
	if music.stopping {
		return
	}
	music.stopping = music.playing
	go stopPlayBack()
	if music.playing {
		<-music.stopped
	}
	music.stopping = false
	fmt.Fprint(w.res, "stopped")
}

// Search sheet names
func (w *Web) serveSearch() {
	type searchRequest struct {
		Keyword string
	}
	request := &searchRequest{}
	w.jsonRequest(request)
	names := sheetSearch(request.Keyword)
	type loadResponse struct {
		Names []string
	}
	response := loadResponse{
		Names: names,
	}
	w.jsonResponse(response)
}

// Loads a sheet
func (w *Web) serveLoadSheet() {
	type loadSheetRequest struct {
		Name string
	}
	request := &loadSheetRequest{}
	w.jsonRequest(request)
	type loadSheetResponse struct {
		Name     string
		Notation string
	}
	sheet := &Sheet{
		Name: filepath.Base(request.Name),
		Dir:  filepath.Dir(request.Name),
	}
	err := sheet.Load()
	if err != nil {
		panic(err)
	}
	response := loadSheetResponse{
		Name:     filepath.Join(sheet.Dir, sheet.Name),
		Notation: sheet.Notation,
	}
	w.jsonResponse(response)
}

// Saves a sheet
func (w *Web) serveSaveSheet() {
	type saveSheetResponse struct {
		Result string
	}
	var result = "Sheet has been saved."
	defer func() {
		response := &saveSheetResponse{
			Result: result,
		}
		w.jsonResponse(response)
	}()
	type saveSheetRequest struct {
		Name     string
		Notation string
	}
	request := &saveSheetRequest{}
	w.jsonRequest(request)
	name := filepath.Base(request.Name)
	id := stringNumber(strings.Split(name, "-")[0])
	if id > 0 && id <= 100 {
		result = "Can't save builtin music sheets with the same name."
		return
	}
	sheet := &Sheet{
		Name:     name,
		Dir:      filepath.Dir(request.Name),
		Notation: request.Notation,
	}
	if len(request.Notation) == 0 && sheet.Exists() {
		// delete sheet
		err := sheet.Delete()
		if err != nil {
			result = fmt.Sprintf("Error: %v", err)
		} else {
			result = "Sheet has been deleted."
		}
	} else {
		// save sheet
		err := sheet.Save()
		if err != nil {
			result = fmt.Sprintf("Error: %v", err)
		} else {
			result = "Sheet has been saved to: " + sheet.Path()
		}
	}
}

// Serves voices page
func (w *Web) serveVoices() {
	type homePage struct {
	}
	data := &homePage{}
	w.execTemplate("header", nil)
	w.execTemplate("/voices", data)
}

// Downloads a voice file
func (w *Web) serveDownloadVoice() {
	type downloadRequest struct {
		Name string
	}
	request := &downloadRequest{}
	w.jsonRequest(request)
	var names []string
	if len(request.Name) > 0 {
		names = append(names, request.Name)
	}
	downloadVoiceFiles(w.res, names)
}

// Execute template with name and data
func (w *Web) execTemplate(name string, data interface{}) {
	if err := w.tmpl.ExecuteTemplate(w.res, name, data); err != nil {
		panic(err)
	}
}

// Reads JSON request
func (w *Web) jsonRequest(request interface{}) {
	requestBody, err := ioutil.ReadAll(w.req.Body)
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(requestBody, request); err != nil {
		panic(err)
	}
}

// Writes JSON response
func (w *Web) jsonResponse(response interface{}) {
	jres, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}
	_, err = w.res.Write(jres)
	if err != nil {
		panic(err)
	}
}

// Downloads natural voices
func downloadVoiceFiles(writer io.Writer, names []string) {
	dir := filepath.Join(beepHomeDir(), "voices")
	if len(names) == 0 {
		names = []string{"piano", "violin"}
	}
	for _, name := range names {
		if !strings.HasSuffix(name, ".zip") {
			name += ".zip"
		}
		url := "http://angiud.com/beep/voices/" + name
		fmt.Fprintf(writer, "Downloading '%s'", url)

		// locate file
		resp, err := http.Head(url)
		if err != nil {
			fmt.Fprintln(writer, " Error:", err)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			fmt.Fprintln(writer, "Error locating file. Status:", resp.StatusCode)
			continue
		}
		fmt.Fprintf(writer, " %s bytes ...\n", numberComma(resp.ContentLength))

		// fetch file
		resp, err = http.Get(url)
		if err != nil {
			fmt.Fprintln(writer, "Error downloading file:", err)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			fmt.Fprintln(writer, "Error downloading. Status:", resp.StatusCode)
			continue
		}
		defer resp.Body.Close()

		// read file
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintln(writer, "Error reading file:", err)
			continue
		}

		// save file
		os.MkdirAll(dir, 0755)
		filename := filepath.Join(dir, name)
		err = ioutil.WriteFile(filename, body, 0644)
		if err != nil {
			fmt.Fprintln(writer, "Error saving file:", err)
			continue
		}

		fmt.Fprintf(writer, "  Saving %s\n", filename)
		beepDefault()
	}
}

// All HTML templates
var webTemplates = `{{define "header"}}
<!DOCTYPE html>
<html>
	<head>
		<link rel="stylesheet" type="text/css" href="css/style.css"/>
		<script src='js/system.js'></script>
	</head>
	<body>
	<div class="header">beep</div>
	<div class="menu">
		<a class="menu" href="/">Home</a> |
		<a class="menu" href="/voices">Voices</a>
	</div>
{{end}}

{{define "/"}}
	<script src='js/home.js'></script>
	<div style='padding:10px'>
		<b>Beep notation:</b> <span id='sheetName'></span><br>
		<textarea id='notation' style='width:99%;height:450px;font-family:monospace;font-size:12px'
			spellcheck='false'>{{.Demo}}</textarea>
		<div style='padding-top:6px'>
			<a id='newSheet' class='button' href='javascript:;'>New Sheet</a>
			<a id='play' class='button' href='javascript:;'>Play</a>
			<a id='stop' class='button' href='javascript:;'>Stop</a>
			<a id='save' class='button' href='javascript:;'>Save</a>
			<a id='load' class='button' href='javascript:;'>Load</a>
			<input id='search' title='Search' style='width:100px;margin-left:5px'>
		</div>
		<div id='result' style='padding-top:10px'>
		</div>
	</div>
	</body>
</html>
{{end}}

{{define "/voices"}}
	<script src='js/voices.js'></script>
	<div style='padding:10px'>
		<div style='padding-bottom:10px'><b>Natural Voices:</b></div>
		Piano <a href='javascript:;' class='link' onclick="downloadVoice('piano')">download</a> (13MB)<br>
		Violin <a href='javascript:;' class='link' onclick="downloadVoice('violin')">download</a> (6.9MB)<br>
		<div id='result' style='padding-top:10px'></div>
	</div>
	</body>
</html>
{{end}}

{{define "pageNotFound"}}
	<div style='padding:10px'>
		{{.}} - page not found
	</div>
	</body>
</html>
{{end}}
`

// Files served by the web server
var webFileMap = map[string]string{
	"/css/style.css": `body {margin:0px;font-size:13px;font-family:arial,sans-serif}
div, td, span, input {font-weight:12px;font-family:arial,sans-serif}
div.header {padding:10px;text-shadow:1px 1px #000;font-size:16px;
	font-weight:bold;background:#3456ab;color:white}
div.menu {padding:4px;padding-left:10px;text-shadow:1px 1px #333;font-size:13px;
	font-weight:bold;background:#456abc;color:white}
a {text-decoration:none;color:black}
a:hover {color:blue}
a.menu {color:white;margin-right:4px}
a.menu:hover {color:yellow}
a.link {color:blue}
a.link:hover {color:red}
a.button {padding:4px 8px 4px 8px;text-shadow:0px 0px #333;font-size:13px;
	font-weight:bold;background:#3456ab;color:white;border-radius:3px}
a.button:hover {color:yellow;box-shadow:1px 1px #999}
a.item {font-size:13px;color:white;}
a.item:hover {color:yellow}
div.item {display:inline-block;font-size:13px;padding:2px;border-bottom:1px dotted #333}
`,
	"/js/system.js": `
function init() {
	window.ids = {}
	var all = document.getElementsByTagName('*')
	for (i=0; i<all.length; i++) {
		var id = all[i].id
		if (id) {
			ids[id] = all[i]
		}
	}
}
function Ajax() {
	if (window.XMLHttpRequest) {
		this.ajax = new XMLHttpRequest()
	} else if (window.ActiveXObject) {
		this.ajax = new ActiveXObject('Microsoft.XMLHTTP')
	}
	if (!this.ajax) {
		alert('No Ajax support is available.')
		return
	}
	this.send = function(url, data) {
		var ajax = this
		this.ajax.onreadystatechange = function() {
			if (this.readyState != 4)
				return
			if (this.status != 200) {
				alert('Request failed.\n\n'+
					'Status Code:'+ this.status +'\n'+
					'Status:'+ this.statusText +'\n'+
					'Response:'+ this.responseText)
				return
			}
			if (ajax.onready)
				ajax.onready(this.responseText)
		}
		this.ajax.open('POST', url, true)
		this.ajax.send(data)
	}
	this.jsonResp = function() {
		var resp = this.ajax.responseText
		if (resp.length == '')
			return null
		if (resp[0] != '{') {
			alert(resp)
		} else {
			try {
				eval('var jres = '+ resp)
				return jres
			} catch (e) {
				alert('JSON Error:\n\n'+ e)
			}
		}
		return null
	}
}
`,
	"/js/home.js": `
window.onload = function() {
	init()
	ids.newSheet.onclick = newSheet
	ids.play.onclick = play
	ids.stop.onclick = stop
	ids.load.onclick = load
	ids.save.onclick = saveSheet
	ids.search.onkeypress = searchPress
	ids.search.onfocus = searchFocus
	ids.search.onblur = searchFocus
	ids.search.onfocus()
}
function newSheet() {
	ids.sheetName.innerText = ''
	ids.notation.value = ''
	ids.result.innerHTML = ''
	ids.notation.focus()
}
function play() {
	if (ids.play.innerText != 'Play') {
		return
	}
	var ajax = new Ajax
	ajax.onready = function(data) {
		reset()
	}
	var data = {
		'Notation': ids.notation.value
	}
	ids.play.innerHTML = 'Play &nbsp;&#9654;'
	ids.stop.innerHTML = 'Stop &nbsp;&#9726;'
	ajax.send('/play', JSON.stringify(data))
}
function load() {
	search('')
}
function searchFocus() {
	if (this.value) {
		this.style.color = ''
		if (this.value == this.title)
			this.value = ''
	} else {
		this.style.color = '#aaa'
		this.value = this.title
	}
}
function searchPress(e) {
	e = e || event
	if (e.keyCode == 13) {
		search(this.value)
	}
}
function search(keyword) {
	var ajax = new Ajax
	ajax.onready = function(data) {
		var h = []
		var jres = this.jsonResp()
		if (jres.Names) {
			var names = jres.Names 
			for (i=0; i<names.length; i++) {
				h.push("<div class='item'>")
				h.push("<a href='javascript:;' class='link' onclick='loadSheet(this.innerText)'>")
				h.push(names[i])
				h.push("</a></div><br>\n")
			}
		} else {
			h.push('No matches found.')
		}
		ids.result.innerHTML = h.join('')
	}
	var data = {
		'Keyword': keyword
	}
	ajax.send('/search', JSON.stringify(data))
}
function loadSheet(name) {
	var ajax = new Ajax
	ajax.onready = function(data) {
		var jres = this.jsonResp()
		ids.sheetName.innerHTML = jres.Name
		ids.notation.value = jres.Notation
		document.body.scrollTop = 0
	}
	var data = {
		'Name': name
	}
	ajax.send('/loadSheet', JSON.stringify(data))
}
function saveNewSheet(elem) {
	ids.sheetName.innerText = elem.previousSibling.value
	saveSheet()
}
function saveSheet() {
	var name = ids.sheetName.innerText
	if (!name) {
		var h = []
		h.push("File Path: <input style='width:350px;margin-right:5px' value='directory/filename.txt'>")
		h.push("<a href='javascript:;' onclick='saveNewSheet(this)' class='button'>Save</a>")
		ids.result.innerHTML = h.join('')
		ids.result.childNodes[1].focus()
		return
	}
	var ajax = new Ajax
	ajax.onready = function(data) {
		var jres = this.jsonResp()
		ids.result.innerHTML = jres.Result
	}
	var data = {
		'Name': ids.sheetName.innerText,
		'Notation': ids.notation.value
	}
	ajax.send('/saveSheet', JSON.stringify(data))
}
function reset() {
	ids.play.innerHTML = 'Play'
	ids.stop.innerHTML = 'Stop'
}
function stop() {
	var ajax = new Ajax
	ajax.onready = function(data) {
		if (data == 'stopped')
			reset()
	}
	ajax.send('/stop', null)
}
`,
	"/js/voices.js": `
window.onload = function() {
	init()
}
function downloadVoice(name) {
	var ajax = new Ajax
	ajax.onready = function(data) {
		ids.result.innerHTML = data.replace(/\n/g, '<br>')
	}
	var data = {
		'Name': name
	}
	ids.result.innerHTML = 'Downloading ...'
	ajax.send('/downloadVoice', JSON.stringify(data))
}
`,
}
