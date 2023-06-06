package main

import "log"
import "os"
import "time"
import "strings"
import "net/http"
import "github.com/radovskyb/watcher"
import "github.com/alexandrevicenzi/go-sse"

type ServerEvent = struct {
    channel string
    message string
}

func main() {
    http.Handle("/style/", FileServer("/style/", "./style/css/"))
    http.Handle("/asset/", CrossOriginFileServer("/asset/", "./asset/"))
    http.Handle("/script/", CrossOriginFileServer("/script/", "./script/"))
    http.Handle("/html/", FileServer("/html/", "./html/"))

    sseServer := StartSseServer()
    defer sseServer.Shutdown()
    sseEvents := make(chan ServerEvent)

    go func() {
        for event := range sseEvents {
            sseServer.SendMessage(event.channel, sse.SimpleMessage(event.message))
        }
    }()
    w := StartWatcher(sseEvents) 
    defer w.Close()
    go func() {
        if err := w.Start(time.Millisecond * 100); err != nil {
            log.Fatalln(err)
        }
    }()

    log.Println("starting server on port localhost:3000")
    err := http.ListenAndServe(":3000", nil)
    if err != nil {
        log.Fatal(err)
    }


}

func FileServer(route string, path string) http.Handler {
    server := http.FileServer(http.Dir(path))
    return http.StripPrefix(route, server)
}

func CrossOriginFileServer(route string, path string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Add("Access-Control-Allow-Origin", "*")
        w.Header().Add("Access-Control-Allow-Methods", "GET, OPTIONS")
        w.Header().Add("Access-Control-Allow-Headers", "Keep-Alive,X-Requested-With,Cache-Control,Content-Type,Last-Event-ID")
        server := http.FileServer(http.Dir(path))
        server = http.StripPrefix(route, server)
        server.ServeHTTP(w, r)
    }
}

func StartSseServer() *sse.Server {
    s := sse.NewServer(&sse.Options{
		// Increase default retry interval to 10s.
		RetryInterval: 10 * 1000,
		// CORS headers
		Headers: map[string]string{
            "Access-Control-Allow-Origin":  "*",
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

func StartWatcher(sseEvents chan ServerEvent) *watcher.Watcher {
    w := watcher.New()
    w.SetMaxEvents(1)
    w.FilterOps(watcher.Write)

    if err := w.AddRecursive("./style/css"); err != nil {
        log.Fatalln(err)
    }

    go func() { 
        for {
			select {
                case event := <-w.Event:	
                    OnFileWrite(event.Path, sseEvents);
                case err := <-w.Error:
                    log.Fatalln(err)
                case <-w.Closed:
                    return
            }
		}
    }()
    


    return w
}

func OnFileWrite(path string, sseEvents chan ServerEvent) {
    if filename, ok := ParseCSSPath(path); ok {
        log.Println(filename)
        sseEvents <- ServerEvent{"/event/style", filename}
    }
}

func ParseCSSPath(path string) (string, bool) {
    if _, filename, ok := strings.Cut(path, "style/css/"); ok {
        if !strings.HasSuffix(filename, ".map") {
            return filename, true
        }
    }
    return "", false
}
