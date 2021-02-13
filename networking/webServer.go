package networking

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	STATUS_ENDPOINT              = "/status"
	HOME_PAGE_ENDPOINT           = "/"
	HOME_PAGE_UI_ASSETS_BASE_DIR = "/ui_assets/"
	RESOURCES_DIR                = "./resources"
)

type WebServer struct {
	port              int
	server            http.Server
	onRequestCallback OnRequestCallback
}

func NewWebServer(port int, onRequestCallback OnRequestCallback) *WebServer {
	ws := WebServer{
		port:              port,
		onRequestCallback: onRequestCallback,
	}

	return &ws
}

func (ws *WebServer) StartServer() {
	m := http.NewServeMux()
	m.HandleFunc(STATUS_ENDPOINT, ws.handleStatusCheckRequest)
	m.HandleFunc(ws.onRequestCallback.GetEndpoint(), ws.handleTaskRequest)

	// handle requests for resources
	m.HandleFunc(HOME_PAGE_ENDPOINT, ws.handleRequestForAsset)

	ws.server = http.Server{Addr: fmt.Sprint(":", ws.port), Handler: m}

	if err := ws.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

func (ws *WebServer) handleRequestForAsset(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	var response []byte

	asset := r.URL.Path
	if asset == HOME_PAGE_ENDPOINT {
		response = readUiAsset(HOME_PAGE_UI_ASSETS_BASE_DIR + "index.html")
	} else {
		response = readUiAsset(asset)
	}

	addContentType(asset, w)

	fmt.Fprint(w, string(response))
}

func readUiAsset(asset string) []byte {
	assetBytes, err := ioutil.ReadFile(RESOURCES_DIR + asset)
	if err != nil {
		fmt.Println(err)
		return []byte{}
	}

	return assetBytes
}

func addContentType(asset string, w http.ResponseWriter) {
	contentType := "text/html"
	if strings.HasSuffix(asset, "js") {
		contentType = "text/javascript"
	} else if strings.HasSuffix(asset, "css") {
		contentType = "text/css"
	}

	w.Header().Set("Content-Type", contentType)
}

func (ws *WebServer) handleTaskRequest(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != ws.onRequestCallback.GetEndpoint() {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	requestBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	responseBytes := ws.onRequestCallback.HandleRequest(requestBytes)

	fmt.Fprint(w, string(responseBytes))
}

func (ws *WebServer) handleStatusCheckRequest(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != STATUS_ENDPOINT {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	fmt.Fprint(w, "Server is alive\n")
}
