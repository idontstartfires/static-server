package main

import "log"
import "os"
import "strings"
import "net/http"
import "github.com/fsnotify/fsnotify"
import "github.com/alexandrevicenzi/go-sse"

type ServerEvent = struct {
    channel string
    message string
}

func main() {
    styleServer := http.FileServer(http.Dir("./style/css"))
    http.Handle("/style/", http.StripPrefix("/style/", styleServer))

    assetServer := http.FileServer(http.Dir("./asset"))
    http.Handle("/asset/", http.StripPrefix("/asset/", assetServer))

    http.HandleFunc("/script/reload.js", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "http://localhost")
        http.ServeFile(w, r, "./script/reload.js")
    })

    http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "./html/test.html")
    })

    
    sseServer := StartSseServer()
    defer sseServer.Shutdown()
    sseEvents := make(chan ServerEvent)

    go func() {
        for {
            event, ok := <- sseEvents
            if !ok {
                return
            }
            sseServer.SendMessage(event.channel, sse.SimpleMessage(event.message))
        }
    }()
    watcher := StartWatcher(sseEvents) 
    defer watcher.Close()
    go func() {
        for {
            event, ok := <-watcher.Events
            if !ok {
                return
            }
            if event.Has(fsnotify.Write) {
                OnFileWrite(event.Name, sseEvents) 
            }
        }
    }()


    log.Println("starting server on port 3000")
    err := http.ListenAndServe(":3000", nil)
    if err != nil {
        log.Fatal(err)
    }


}

func StartSseServer() *sse.Server {
    s := sse.NewServer(&sse.Options{
		// Increase default retry interval to 10s.
		RetryInterval: 10 * 1000,
		// CORS headers
		Headers: map[string]string{
            "Access-Control-Allow-Origin":  "http://localhost",
			"Access-Control-Allow-Methods": "GET, OPTIONS",
			"Access-Control-Allow-Headers": "Keep-Alive,X-Requested-With,Cache-Control,Content-Type,Last-Event-ID",
		},
		// Custom channel name generator
		ChannelNameFunc: func(request *http.Request) string {
			return request.URL.Path
		},
		// Print debug info
		Logger: log.New(os.Stdout, "go-sse: ", log.Ldate|log.Ltime|log.Lshortfile),
	});
    http.Handle("/event/", s)
    return s
}

func StartWatcher(sseEvents chan ServerEvent) *fsnotify.Watcher {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }

    

    err = watcher.Add("./style/css")
    if err != nil {
        log.Fatal(err)
    }
    return watcher
}

func OnFileWrite(path string, sseEvents chan ServerEvent) {
    if filename, ok := ParseCSSPath(path); ok {
        sseEvents <- ServerEvent{"/event/style", filename}
    }
}

func ParseCSSPath(path string) (string, bool) {
    if filename, ok := strings.CutPrefix(path, "style/css/"); ok {
        if !strings.HasSuffix(filename, ".map") {
            return filename, true
        }
    }
    return "", false
}
