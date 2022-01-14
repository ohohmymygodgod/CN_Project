package main

import (
    "bufio"
    "fmt"
    "net"
    "strings"
    "os"
    // "time"
    "io"
)

var question string
var listener net.Listener


type Client struct {
    name       string
    serverReader     *bufio.Reader
    serverWriter     *bufio.Writer
    clientReader     *bufio.Reader
    clientWriter     *bufio.Writer
    clientConn  net.Conn
    serverConn  net.Conn
}


func (client *Client) consoleRead() {
    for {
        message, err := client.serverReader.ReadString('\n')
        message = strip(message)
        checkErr(err)
        if(message == "FINISH") {
            return
        }
        fmt.Println(message)
    }
}

func (client *Client) consoleWrite() {
    in := bufio.NewReader(os.Stdin)
    for {
        message, err := in.ReadString('\n')
        message = strip(message)
        checkErr(err)
        client.serverWriter.WriteString(message+"\n")
        client.serverWriter.Flush()
        if(message == "exit") {
            return
        }
    }
}

func (client *Client) consoleAskName() {
    fmt.Println("input your username:")
    var name string
    fmt.Scanln(&name)
    client.name = name
    client.serverWriter.WriteString(name+"\n")
    client.serverWriter.Flush()
}

func (client *Client) consoleReadOption() (string, string) {
    var option string
    fmt.Scanln(&option)
    client.serverWriter.WriteString(option+"\n")
    client.serverWriter.Flush()
    if(option == "exit") {
        return "exit", option
    }
    res, err := client.serverReader.ReadString('\n')
    res = strip(res)
    checkErr(err)
    fmt.Println(res)
    return res, option
}


func (client *Client) consoleWorkFlow() {
    client.consoleAskName()
    for {
        fmt.Println(question)
        res, option := client.consoleReadOption()
        switch option {
            case "1":
            case "2":
                if(res == "No other users.") {
                    break
                }
                client.consoleReadOption()
            case "3":
                if(res == "No friends.") {
                    break
                }
                client.consoleReadOption()
            case "4":
                if(res == "No friends.") {
                    break
                }
                fmt.Scanln(&option)
                client.serverWriter.WriteString(option+"\n")
                client.serverWriter.Flush()
                for {
                    res, err := client.serverReader.ReadString('\n')
                    res = strip(res)
                    checkErr(err)
                    if(res != "START") {
                        fmt.Println(res)
                    }else{
                        break
                    }
                }
                go client.consoleRead()
                client.consoleWrite()
            case "exit":
                client.exit()
        }
    }
}

func (client *Client) parse (input string) (string, string, string, map[string]string) {
    if(input == ""){
        return "", "", "", make(map[string]string)
    }
    params := make(map[string]string)
    tmp := strings.Split(input, " ")
    method := tmp[0]
    version := tmp[2]
    tmp = strings.Split(tmp[1], "?")
    path := tmp[0]
    if(len(tmp) > 1){
        tmp = strings.Split(tmp[1], "&")
        for _, item := range tmp {
            param := strings.Split(item, "=")
            params[param[0]] = param[1]
        }
    }
    fmt.Println(method, path, version, params)
    return method, path, version, params
}

func (client *Client) readHttp() (string, string, string, map[string]string)  {
    count := 0
    var ret string
    for {
        line, err := client.clientReader.ReadString('\r')
        line = strip(line)
        if(err != io.EOF){
            checkErr(err)
        }
        if line == "" {
            return client.parse(ret)
        }
        fmt.Println(line)
        if(count == 0) {
            ret = line
        }
        count++
    }
}

func (client *Client) sendIndex(version string) {
    s, err := os.ReadFile("./template/index.html")
    checkErr(err)
    header := fmt.Sprintf("%s 200 OK\r\nAccept-Ranges: bytes\r\nContent-Length: %d\r\nContent-Type: text/html; charset=UTF-8\r\nConnection: close\r\n\r\n", version, len(s))
    client.clientWriter.WriteString(header+string(s)+"\r\n")
    client.clientWriter.Flush()
}


func (client *Client) webWorkFlow() {
    for {
        method, path, version, params := client.readHttp()
        if(method == ""){
            // time.Sleep(1*time.Second)
            // fmt.Println("here")
            continue
        }
        fmt.Println(method, path, version, params)
        switch method {
            case "GET":
                client.sendIndex(version)
        }

    }
}

func NewClient(connection net.Conn) *Client {
    serverConn, err := net.Dial("tcp", "127.0.0.1:8888")
    checkErr(err)
    serverWriter := bufio.NewWriter(serverConn)
    serverReader := bufio.NewReader(serverConn)
    clientWriter := bufio.NewWriter(connection)
    clientReader := bufio.NewReader(connection)

    client := &Client{
        clientConn: connection,
        serverConn: serverConn,
        serverReader: serverReader,
        serverWriter: serverWriter,
        clientReader: clientReader,
        clientWriter: clientWriter,
    }
    return client
}

func console() {
    tmp := new(net.Conn)
    client := NewClient(*tmp)
    client.consoleWorkFlow()
}

func web() {
    listener, _ = net.Listen("tcp", ":80")
    conn, err := listener.Accept()
    checkErr(err)
    client := NewClient(conn)
    client.webWorkFlow()
}

func main() {
    mode := os.Args[1]
    fmt.Println(mode)
    question = "Home\n (1) List all friends\n (2) Add friend\n (3) Delete friend\n (4) Choose a chat room"
    switch mode {
    case "console":
        console()
    case "web":
        web()
    }
}

func (client *Client)exit() {
    client.serverConn.Close()
    if(client.clientConn != nil) {
        client.clientConn.Close()
    }
    os.Exit(1)
}


func strip(s string) string {
	s = strings.Replace(s, "\n", "", -1)
	s = strings.Replace(s, "\r", "", -1)
	return s
}

func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}