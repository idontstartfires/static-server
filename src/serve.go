package main

import "log"
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

    scriptServer := http.FileServer(http.Dir("./script"))
    http.Handle("/script/", http.StripPrefix("/script/", scriptServer))

    http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "./html/test.html")
    })

    
    sseEvents := CreateServerSentEventChannel()
    StartWatcher(sseEvents) 

    log.Println("starting server on port 3000")
    err := http.ListenAndServe(":3000", nil)
    if err != nil {
        log.Fatal(err)
    }


}

func CreateServerSentEventChannel() chan ServerEvent {
    s := sse.NewServer(nil)
    defer s.Shutdown()
    http.Handle("/event/", s)
    sseEvents := make(chan ServerEvent)

    go func() {
        for {
            event, ok := <- sseEvents
            if !ok {
                return
            }
            s.SendMessage(event.channel, sse.SimpleMessage(event.message))
        }
    }()

    return sseEvents
}

func StartWatcher(sseEvents chan ServerEvent) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }
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

    err = watcher.Add("./style/css")
    if err != nil {
        log.Fatal(err)
    }
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
