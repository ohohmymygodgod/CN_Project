package main

import (
    "bufio"
    "fmt"
    "net"
    "strings"
    "os"
    "io"
    "time"
)

var question string


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

func (client *Client) parse (input string) {
    // var path string
    // var param map[string]string
    tmp := strings.Split(input, "?")
    fmt.Println(tmp)
}

func (client *Client) get(req []string) {
    client.parse(req[1])
}

func (client *Client) readHttp() {
    count := 0
    fmt.Println("here")
    for{
        http, err := client.clientReader.ReadString('\n')
        if(err != nil){
            if(err != io.EOF){
                checkErr(err)
            }else{
                time.Sleep(5*time.Second)
            }
        }
        if(count == 0){
            req := strings.Split(http, " ")
            if(req[0] == "GET") {
                client.get(req)
            }
        }
        if(http == "\r\n"){
            break
        }
        count++
    }
}

func (client *Client) webAskName() {
    client.readHttp()
    s, err := os.ReadFile("./index.html")
    checkErr(err)
    header := fmt.Sprintf("HTTP/1.1 200 OK\r\nAccept-Ranges: bytes\r\nContent-Length: %d\r\nContent-Type: text/html; charset=UTF-8\n\n", len(s))
    fmt.Print(header+string(s)+"\r\n")
    client.clientWriter.WriteString(header+string(s)+"\r\n")
    client.clientWriter.Flush()
    
    client.readHttp()
}

func (client *Client) webWorkFlow() {
    client.webAskName()
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
    listener, _ := net.Listen("tcp", ":80")
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
	return s
}

func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}