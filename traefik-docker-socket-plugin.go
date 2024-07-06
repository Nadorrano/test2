package main

import (
    "context"
    "io"
    "net"
    "net/http"
)

type Config struct {
    SocketPath string `json:"socketPath,omitempty"`
}

func CreateConfig() *Config {
    return &Config{}
}

type DockerSocketPlugin struct {
    next      http.Handler
    name      string
    socketPath string
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
    return &DockerSocketPlugin{
        next:      next,
        name:      name,
        socketPath: config.SocketPath,
    }, nil
}

func (d *DockerSocketPlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
    client := &http.Client{
        Transport: &http.Transport{
            DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
                return net.Dial("unix", d.socketPath)
            },
        },
    }

    req.URL.Scheme = "http"
    req.URL.Host = "docker"

    resp, err := client.Do(req)
    if err != nil {
        http.Error(rw, err.Error(), http.StatusBadGateway)
        return
    }
    defer resp.Body.Close()

    rw.WriteHeader(resp.StatusCode)
    io.Copy(rw, resp.Body)
}

