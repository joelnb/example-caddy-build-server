package main

import (
    "fmt"
    "io/ioutil"
    "log"
    "math/rand"
    "net/http"
    "os"
    "os/exec"
    "sync"
)

// curl -v "localhost:8081/api/download?os=linux&arch=amd64&p=github.com/caddy-dns/lego-deprecated" --output caddy

func main() {
    var currentWork sync.Map

    http.HandleFunc("/api/download", func(w http.ResponseWriter, r *http.Request) {
        idempotency := r.URL.Query().Get("idempotency")
        if idempotency == "" {
            idempotency = fmt.Sprintf("%d", rand.Int())
        }

        _, loaded := currentWork.LoadOrStore(idempotency, true)
        if loaded {
            fmt.Printf("%s: Already doing work for this idempotency value!\n", idempotency)
            w.WriteHeader(http.StatusBadRequest)
            fmt.Fprintf(w, "400 - Build already running for %s!", idempotency)
            return
        }

        outPath := fmt.Sprintf("/tmp/xcaddy-build-%s", idempotency)
        command := "xcaddy"
        cmdArgs := []string{"build", "--output", outPath}

        plugins := r.URL.Query()["p"]
        for _, plugin := range plugins {
            cmdArgs = append(cmdArgs, "--with", plugin)
        }

        goos := fmt.Sprintf("GOOS=%s", r.URL.Query().Get("os"))
        goarch := fmt.Sprintf("GOARCH=%s", r.URL.Query().Get("arch"))

        fmt.Printf("%s: Running command '%s' with args: %s\n", idempotency, command, cmdArgs)
        fmt.Printf("%s: Applying env vars: %s %s\n", idempotency, goos, goarch)

        cmd := exec.Command(command, cmdArgs...)
        cmd.Env = append(os.Environ(), goos, goarch)
        out, err := cmd.Output()

        fmt.Printf("%s: Command output: %s\n", idempotency, out)

        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            fmt.Fprintf(w, "500 - Build failed for %s!", idempotency)
        } else {
            fileBytes, err := ioutil.ReadFile(outPath)
            if err != nil {
                fmt.Printf("%s: Unable to read built file: %s", idempotency, err)

                w.WriteHeader(http.StatusInternalServerError)
                fmt.Fprintf(w, "500 - Build failed for %s!", idempotency)

                e := os.Remove(outPath)
                if e != nil {
                    fmt.Printf("%s: Unable to cleanup built file: %s", idempotency, err)
                    return
                }

                currentWork.Delete(idempotency)

                return
            }

            w.WriteHeader(http.StatusOK)
            w.Header().Set("Content-Type", "application/octet-stream")
            w.Write(fileBytes)
        }

        e := os.Remove(outPath)
        if e != nil {
            fmt.Printf("%s: Unable to cleanup built file: %s", idempotency, err)
            return
        }

        currentWork.Delete(idempotency)
    })

    log.Print("Starting caddy build server")
    log.Fatal(http.ListenAndServe(":8081", nil))
}
