package main

import (
    "log"
    "path/filepath"
    "time"

    "github.com/fsnotify/fsnotify"
)

func watchFiles(path string) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }
    defer watcher.Close()

    done := make(chan bool)

    go func() {
        for {
            select {
            case event, ok := <-watcher.Events:
                if !ok {
                    return
                }
                log.Println("File event:", event)
                if event.Op&fsnotify.Create == fsnotify.Create {
                    ext := filepath.Ext(event.Name)
                    if ext == ".php" {
                        eventData := HTTPEvent{
                            Timestamp:   time.Now(),
                            Description: "Possible WebShell upload detected",
                        }
                        sendHTTPEvent(eventData)
                    }
                }
            case err, ok := <-watcher.Errors:
                if !ok {
                    return
                }
                log.Println("Watcher error:", err)
            }
        }
    }()

    err = watcher.Add(path)
    if err != nil {
        log.Fatal(err)
    }
    <-done
}
