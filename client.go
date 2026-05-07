package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
    "strings"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: go run client.go <username>")
        return
    }

    username := os.Args[1]
    conn, err := net.Dial("tcp", "localhost:8989")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    defer conn.Close()

    // Читаем сообщения от сервера
    go func() {
        scanner := bufio.NewScanner(conn)
        for scanner.Scan() {
            fmt.Println(scanner.Text())
        }
    }()

    // Отправляем имя
    conn.Write([]byte(username + "\n"))

    // Отправляем сообщения
    reader := bufio.NewReader(os.Stdin)
    for {
        text, _ := reader.ReadString('\n')
        text = strings.TrimSpace(text)
        if text != "" {
            conn.Write([]byte(text + "\n"))
        }
    }
}