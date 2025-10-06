package beep

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

// Web params
type Web struct {
	music *Music
	tmpl  *template.Template
}

// StartWebServer starts beep web server
func StartWebServer(music *Music, address string) {
	var err error
	var ip string
	if len(address) == 0 {
		ip = "127.0.0.1"
	}
	parts := strings.Split(address, ":")
	if len(parts[0]) > 0 {
		ip = parts[0]
	}
	port := 4444
	if len(parts) > 1 && len(parts[1]) > 0 {
		port, err = strconv.Atoi(parts[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, "Invalid port number:", parts[1])
			os.Exit(1)
		}
	}

	bind := fmt.Sprintf("%s:%d", ip, port)
	web := NewWeb(music)
	music.quietMode = true
	if len(ip) == 0 {
		fmt.Println("Listening on port", port)
	} else {
		fmt.Printf("Listening on http://%s/\n", bind)
	}
	err = http.ListenAndServe(bind, web)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to start web server:", err)
		os.Exit(1)
	}
}

// NewWeb returns new handler
func NewWeb(music *Music) *Web {
	w := &Web{
		music: music,
		tmpl:  template.Must(template.New("tmpl").Parse(webTemplates)),
	}
	return w
}

func (w *Web) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	defer func() {
		if obj := recover(); obj != nil {
			format := "\" ><pre>\nError: %s\nStack:%s\n</pre>"
			fmt.Fprintf(res, format, obj, debug.Stack())
		}
	}()
	path := req.URL.Path
	if file, found := webFileMap[path]; found {
		if strings.HasSuffix(path, ".css") {
			res.Header().Add("Content-Type", "text/css")
		}
		if strings.HasSuffix(path, ".js") {
			res.Header().Add("Content-Type", "text/javascript")
		}
		fmt.Fprint(res, file)
		return
	}
	defer req.Body.Close()
	switch path {
	case "/":
		w.serveHome(res, req)
	case "/play":
		w.servePlay(res, req)
	case "/stop":
		w.serveStop(res, req)
	case "/search":
		w.serveSearch(res, req)
	case "/loadSheet":
		w.serveLoadSheet(res, req)
	case "/saveSheet":
		w.serveSaveSheet(res, req)
	case "/voices":
		w.serveVoices(res, req)
	case "/downloadVoice":
		w.serveDownloadVoice(res, req)
	case "/exportWave":
		w.serveExportWave(res, req)
	default:
		w.execTemplate("header", nil, res)
		w.execTemplate("pageNotFound", req.URL.Path, res)
	}
}

// Serves home page
func (w *Web) serveHome(res http.ResponseWriter, req *http.Request) {
	type homePage struct {
		Demo string
	}
	data := &homePage{
		Demo: DemoMusic,
	}
	w.execTemplate("header", nil, res)
	w.execTemplate("/", data, res)
}

// Playback
func (w *Web) servePlay(res http.ResponseWriter, req *http.Request) {
	if w.music.playing {
		return
	}
	if w.music.stopping {
		return
	}
	type playRequest struct {
		Notation string
	}
	request := &playRequest{}
	w.jsonRequest(request, req)
	InitSoundDevice()
	notation := bytes.NewBuffer([]byte(request.Notation))
	reader := bufio.NewReader(notation)
	go w.music.Play(reader, 100)
	w.music.Wait()
}

// Stops playback
func (w *Web) serveStop(res http.ResponseWriter, req *http.Request) {
	if w.music.stopping {
		return
	}
	w.music.stopping = w.music.playing
	go StopPlayBack()
	if w.music.playing {
		<-w.music.stopped // wait until Music.Play() exits
	}
	FlushSoundBuffer()
}

// Search sheet names
func (w *Web) serveSearch(res http.ResponseWriter, req *http.Request) {
	type searchRequest struct {
		Keyword string
	}
	request := &searchRequest{}
	w.jsonRequest(request, req)
	names := sheetSearch(request.Keyword)
	type loadResponse struct {
		Names []string
	}
	response := loadResponse{
		Names: names,
	}
	w.jsonResponse(response, res)
}

// Loads a sheet
func (w *Web) serveLoadSheet(res http.ResponseWriter, req *http.Request) {
	type loadSheetRequest struct {
		Name string
	}
	request := &loadSheetRequest{}
	w.jsonRequest(request, req)
	type loadSheetResponse struct {
		Name     string
		URL      string
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
		URL:      sheet.URL,
		Notation: sheet.Notation,
	}
	w.jsonResponse(response, res)
}

// Saves a sheet
func (w *Web) serveSaveSheet(res http.ResponseWriter, req *http.Request) {
	type saveSheetResponse struct {
		Result string
	}
	var result = "Sheet has been saved."
	defer func() {
		response := &saveSheetResponse{
			Result: result,
		}
		w.jsonResponse(response, res)
	}()
	type saveSheetRequest struct {
		Name     string
		Notation string
	}
	request := &saveSheetRequest{}
	w.jsonRequest(request, req)
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
func (w *Web) serveVoices(res http.ResponseWriter, req *http.Request) {
	type homePage struct {
	}
	data := &homePage{}
	w.execTemplate("header", nil, res)
	w.execTemplate("/voices", data, res)
}

// Downloads a voice file
func (w *Web) serveDownloadVoice(res http.ResponseWriter, req *http.Request) {
	type downloadRequest struct {
		Name string
	}
	request := &downloadRequest{}
	w.jsonRequest(request, req)
	var names []string
	if len(request.Name) > 0 {
		names = append(names, request.Name)
	}
	DownloadVoiceFiles(w.music, res, names)
}

// Export to WAV file
func (w *Web) serveExportWave(res http.ResponseWriter, req *http.Request) {
	defer func() {
		w.music.output = ""
	}()
	type exportWaveRequest struct {
		Output   string
		Notation string
	}
	request := &exportWaveRequest{}
	w.jsonRequest(request, req)

	notation := bytes.NewBuffer([]byte(request.Notation))
	reader := bufio.NewReader(notation)
	w.music.output = filepath.Join(HomeDir(), "export", request.Output)
	os.MkdirAll(filepath.Dir(w.music.output), 0755)
	go w.music.Play(reader, 100)
	w.music.Wait()

	type exportWaveResponse struct {
		Result string
	}
	response := exportWaveResponse{
		Result: "WAV file has been save to: " + w.music.output,
	}
	w.jsonResponse(response, res)
}

// Execute template with name and data
func (w *Web) execTemplate(name string, data interface{}, res http.ResponseWriter) {
	if err := w.tmpl.ExecuteTemplate(res, name, data); err != nil {
		panic(err)
	}
}

// Reads JSON request
func (w *Web) jsonRequest(request interface{}, req *http.Request) {
	requestBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(requestBody, request); err != nil {
		panic(err)
	}
}

// Writes JSON response
func (w *Web) jsonResponse(response interface{}, res http.ResponseWriter) {
	jres, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}
	_, err = res.Write(jres)
	if err != nil {
		panic(err)
	}
}

// DownloadVoiceFiles downloads natural voice files
func DownloadVoiceFiles(music *Music, writer io.Writer, names []string) {
	dir := filepath.Join(HomeDir(), "voices")
	if len(names) == 0 {
		names = []string{"piano", "violin"}
	}
	for _, name := range names {
		if !strings.HasSuffix(name, ".zip") {
			name += ".zip"
		}
		url := "http://bmrust.com/dl/beep/voices/" + name
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
	}

	// reload voices
	music.piano = NewPiano()
	music.violin = NewViolin()
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
	<div class="header" ondblclick="this.style.display='none'">beep</div>
	<div class="menu">
		<a class="menu" href="/">Home</a> |
		<a class="menu" href="/voices">Voices</a>
	</div>
{{end}}

{{define "/"}}
	<script src='js/home.js'></script>
	<div style='padding:10px'>
		<b>Beep notation:</b> <span id='sheetName'></span>&nbsp;
		<a id='sheetUrl' target='_blank' style='color:blue'></a><br>
		<textarea id='notation' style='width:99%;height:450px;font-family:monospace;font-size:12px'
			spellcheck='false'>{{.Demo}}</textarea>
		<div style='padding-top:6px'>
			<a id='newSheet' class='button' href='javascript:;'>New Sheet</a>
			<a id='play' class='button' href='javascript:;'>Play</a>
			<a id='stop' class='button' href='javascript:;'>Stop</a>
			<a id='save' class='button' href='javascript:;'>Save</a>
			<a id='load' class='button' href='javascript:;'>Load</a>
			<a id='exportWave' class='button' href='javascript:;'>Export</a>
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
	ids.exportWave.onclick = exportPath
	ids.search.onfocus()
}
function newSheet() {
	ids.sheetName.innerText = ''
	ids.sheetUrl.href = ''
	ids.sheetUrl.innerText = ''
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
	ids.result.innerHTML = 'Searching ...'
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
		setTimeout(function() {ids.result.innerHTML = h.join('')}, 200)
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
		ids.sheetUrl.href = jres.URL
		ids.sheetUrl.innerText = jres.URL ? 'Music sheet' : ''
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
	ids.result.innerHTML = 'Saving ...'
	var ajax = new Ajax
	ajax.onready = function(data) {
		var jres = this.jsonResp()
		setTimeout(function() {ids.result.innerHTML = jres.Result}, 200)
	}
	var data = {
		'Name': ids.sheetName.innerText,
		'Notation': ids.notation.value
	}
	ajax.send('/saveSheet', JSON.stringify(data))
}
function exportPath() {
	var h = []
	h.push("File Path: <input style='width:350px;margin-right:5px' value='output.wav'>")
	h.push("<a href='javascript:;' onclick='exportWave(this)' class='button'>Export</a>")
	ids.result.innerHTML = h.join('')
	ids.result.childNodes[1].focus()
}
function exportWave(elem) {
	var output = elem.previousSibling.value
	if (!output) return
	ids.result.innerHTML = 'Exporting ...'
	var ajax = new Ajax
	ajax.onready = function(data) {
		var jres = this.jsonResp()
		ids.result.innerHTML = jres.Result
	}
	var data = {
		'Output': output,
		'Notation': ids.notation.value
	}
	ajax.send('/exportWave', JSON.stringify(data))
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
