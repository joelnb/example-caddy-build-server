package main

import (
    "fmt"
    "log"
    "math/rand"
    "net/http"
    "os"
    "os/exec"
    "sync"
)

var currentWork sync.Map

func SafeRemove(path string) error {
    if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
        return err
    }

    return nil
}

func Write500(w http.ResponseWriter) {
    w.WriteHeader(http.StatusInternalServerError)
    http.Error(w, "Error running build", http.StatusInternalServerError)
}

func HandleXCaddyDownload(w http.ResponseWriter, r *http.Request) {
    idempotency := r.URL.Query().Get("idempotency")
    if idempotency == "" {
        idempotency = fmt.Sprintf("%d", rand.Int())
    }

    _, loaded := currentWork.LoadOrStore(idempotency, true)
    if loaded {
        log.Printf("%s: Already doing work for this idempotency value!\n", idempotency)
        http.Error(w, fmt.Sprintf("400 - Build already running for %s!", idempotency), http.StatusBadRequest)
        return
    }

    outPath := fmt.Sprintf("/tmp/xcaddy-build-%s", idempotency)

    if err := SafeRemove(outPath); err != nil {
        log.Printf("%s: Unable to cleanup unexpected existing file: %s", idempotency, err)
        return
    }

    command := "xcaddy"
    cmdArgs := []string{"build", "--output", outPath}

    plugins := r.URL.Query()["p"]
    for _, plugin := range plugins {
        cmdArgs = append(cmdArgs, "--with", plugin)
    }

    goos := fmt.Sprintf("GOOS=%s", r.URL.Query().Get("os"))
    goarch := fmt.Sprintf("GOARCH=%s", r.URL.Query().Get("arch"))

    log.Printf("%s: Running command '%s' with args: %s", idempotency, command, cmdArgs)
    log.Printf("%s: Applying env vars: %s %s", idempotency, goos, goarch)

    cmd := exec.Command(command, cmdArgs...)
    cmd.Env = append(os.Environ(), goos, goarch)
    out, err := cmd.CombinedOutput()

    log.Printf("%s: Command output: %s", idempotency, out)

    defer func() {
        currentWork.Delete(idempotency)
    }()

    if err != nil {
        Write500(w)
    } else {
        defer func() {
            if err := SafeRemove(outPath); err != nil {
                log.Printf("%s: Unable to cleanup built file: %s", idempotency, err)
            }
        }()

        fileBytes, err := os.ReadFile(outPath)
        if err != nil {
            log.Printf("%s: Unable to read built file: %s", idempotency, err)
            Write500(w)
        } else {
            w.WriteHeader(http.StatusOK)
            w.Header().Set("Content-Type", "application/octet-stream")

            if _, err := w.Write(fileBytes); err != nil {
                log.Printf("%s: Writing file to client failed: %v", idempotency, err)
            }
        }
    }
}

func main() {
    http.HandleFunc("/api/download", HandleXCaddyDownload)

    log.Print("Starting caddy build server")
    log.Fatal(http.ListenAndServe(":8081", nil))
}
