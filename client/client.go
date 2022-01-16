package main

import (
    "bufio"
    "fmt"
    "net"
    "strings"
    "os"
    // "io/ioutil"
    // "time"
    // "io"
)

var question string
var listener net.Listener
var s string
var data []byte
var IP, port string

type Client struct {
    file chan string
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

func (client *Client) parseParam (input string) map[string]string {
    params := make(map[string]string)
    tmp := strings.Split(input, "&")
    for _, item := range tmp {
        param := strings.Split(item, "=")
        params[param[0]] = param[1]
    }
    return params
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
    if(len(tmp) > 1 && tmp[1] != ""){
        params = client.parseParam(tmp[1])
    }
    fmt.Println(method, path, version, params)
    return method, path, version, params
}

func (client *Client) writeFile(n int, input string) []byte {
    tmp := []byte(input)
    buf := make([]byte, 4096)
    for (n == 4096) {
        l, err := client.clientConn.Read(buf)
        n = l
        checkErr(err)
        tmp = append(tmp, buf[:l]...)
    }
    tmp = append(tmp, buf...)
    return tmp
}

func (client *Client) readHttp() (string, string, string, map[string]string)  {
    count := 0
    var method string
    var path string
    var version string
    params := make(map[string]string)
    buf := make([]byte, 4096)
    n, err := client.clientConn.Read(buf)
    checkErr(err)
	req := strings.Split(string(buf[:n]), "\r\n")
    l := len(req)
    for i, _ := range req {
        // fmt.Println(req[i])
        if(count == 0) {
            method, path, version, params = client.parse(req[i])
        }
        count++
        if(path == "/file"){
            fmt.Println(n, len(req[l-1]))
            data = client.writeFile(n, req[l-1])
            return method, path, version, make(map[string]string)
        }
        if(i == l-1) {
            if(method == "POST"){
                params = client.parseParam(req[i])
            }
            return method, path, version, params
        }
    }
    return method, path, version, params
}

func (client *Client) sendIndex(version string) {
    str, err := os.ReadFile("./template/index.html")
    checkErr(err)
    header := fmt.Sprintf("%s 200 OK\r\nAccept-Ranges: bytes\r\nContent-Length: %d\r\nContent-Type: text/html; charset=UTF-8\r\nConnection: close\r\n\r\n", version, len(str))
    client.clientWriter.WriteString(header+string(str)+"\r\n")
    client.clientWriter.Flush()
}


func (client *Client) sendQuestion(version string) {
    str, err := os.ReadFile("./template/home.html")
    checkErr(err)
    header := fmt.Sprintf("%s 200 OK\r\nAccept-Ranges: bytes\r\nContent-Length: %d\r\nContent-Type: text/html; charset=UTF-8\r\nConnection: close\r\n\r\n", version, len(str))
    client.clientWriter.WriteString(header+string(str)+"\r\n")
    client.clientWriter.Flush()
}

func (client *Client) sendOKHtml (res string, version string) {
    prefix := "<html><head><meta charset='utf-8'><title>Chat Box</title></head><body><p>"
	suffix := "</p><form action=\"/home\" method=\"get\" ><label for=\"ok\"></label><input type=\"submit\" value='OK'></form></body></html>"
    str := prefix + res  + suffix
	header := fmt.Sprintf("%s 200 OK\r\nAccept-Ranges: bytes\r\nContent-Length: %d\r\nContent-Type: text/html; charset=UTF-8\r\nConnection: close\r\n\r\n", version, len(str))
    client.clientWriter.WriteString(header+string(str)+"\r\n")
    client.clientWriter.Flush()
}

func (client *Client) sendChooseHtml(res string, version string) {
    prefix := "<html><head><meta charset='utf-8'><title>Chat Box</title></head><body><p>"
	suffix := "</p><form action=\"/choose\" method=\"post\" ><label for=\"choose\"></label><input type=\"text\" id=\"choose\" name=\"choose\"/><br><input type=\"submit\" value='submit'></form></body></html>"
    str := prefix + res  + suffix
	header := fmt.Sprintf("%s 200 OK\r\nAccept-Ranges: bytes\r\nContent-Length: %d\r\nContent-Type: text/html; charset=UTF-8\r\nConnection: close\r\n\r\n", version, len(str))
    client.clientWriter.WriteString(header+string(str)+"\r\n")
    client.clientWriter.Flush()
}

func (client *Client) sendChatChooseHtml(res string, version string) {
    prefix := "<html><head><meta charset='utf-8'><title>Chat Box</title></head><body><p>"
	suffix := "</p><form action=\"/chatChoose\" method=\"post\" ><label for=\"choose\"></label><input type=\"text\" id=\"chatChoose\" name=\"chatChoose\"/><br><input type=\"submit\" value='submit'></form></body></html>"
    str := prefix + res  + suffix
	header := fmt.Sprintf("%s 200 OK\r\nAccept-Ranges: bytes\r\nContent-Length: %d\r\nContent-Type: text/html; charset=UTF-8\r\nConnection: close\r\n\r\n", version, len(str))
    client.clientWriter.WriteString(header+string(str)+"\r\n")
    client.clientWriter.Flush()
}

func (client *Client) sendChatHtml(res string, version string) {
    script, err := os.ReadFile("./template/file.html")
    checkErr(err)
    prefix := "<html><head><meta charset='utf-8'><title>Chat Box</title></head><body><p>"
	suffix := "</p><form action=\"/chat\" method=\"post\" ><label for=\"choose\"></label><input type=\"text\" id=\"message\" name=\"message\"/><br><label for=\"file\">Choose a file:</label><br><input type=\"file\" id=\"file\" name=\"file\"><br><input type=\"submit\" id=\"fileUpload\" value='submit'></form>" + string(script) + "</body></html>"
    str := prefix + res  + suffix
	header := fmt.Sprintf("%s 200 OK\r\nAccept-Ranges: bytes\r\nContent-Length: %d\r\nContent-Type: text/html; charset=UTF-8\r\nConnection: close\r\n\r\n", version, len(str))
    client.clientWriter.WriteString(header+string(str)+"\r\n")
    client.clientWriter.Flush()
}

func (client *Client) sendListFriends(res string, version string) {
	client.sendOKHtml(res, version)
}


func (client *Client) sendAddFriend(res string, version string) {
    if(res == "No other users."){
        client.sendOKHtml(res, version)
    }else{
        client.sendChooseHtml(res, version)
    }
}


func (client *Client) sendDeleteFriend(res string, version string) {
    if(res == "No friends."){
        client.sendOKHtml(res, version)
    }else{
        client.sendChooseHtml(res, version)
    }
}

func (client *Client) webRead(version string){
    for {
        message, err := client.serverReader.ReadString('\n')
        message = strip(message)
        checkErr(err)
        if(message == "FINISH") {
            return
        }
        s += message + "<br>"
        client.sendChatHtml(s, version)
    }
}

func (client *Client) sendChat(res string, version string) {
    if(res == "No friends."){
        client.sendOKHtml(res, version)
    }else{
        client.sendChatChooseHtml(res, version)
    }
}

func (client *Client) sendFile (pathArr []string, version string) {
    name := pathArr[2]
    str, err := os.ReadFile("./files/"+name)
    checkErr(err)
    header := fmt.Sprintf("%s 200 OK\r\nAccept-Ranges: bytes\r\nContent-Length: %d\r\nContent-Type: text/html; charset=UTF-8\r\nConnection: close\r\n\r\n", version, len(str))
    client.clientWriter.WriteString(header+string(str)+"\r\n")
    client.clientWriter.Flush()
}


func (client *Client) reconnect() {
    conn, err := listener.Accept()
    checkErr(err)
    fmt.Println("Client from: ", conn.RemoteAddr())
    client.clientConn = conn
    client.clientWriter = bufio.NewWriter(conn)
    client.clientReader = bufio.NewReader(conn)
}

func (client *Client) webWorkFlow(listener net.Listener) {
    for {
        method, path, version, params := client.readHttp()
        fmt.Println("method, path, version, params = ", method, path, version, params)
        pathArr := strings.Split(path, "/")
        if(len(pathArr) > 1){
            path = pathArr[1]
        }else{
            path = pathArr[0]
        }
        switch method {
            case "GET":
                switch path {
                case "":
                    client.sendIndex(version)
                case "home":
                    client.sendQuestion(version)
                case "files":
                    client.sendFile(pathArr, version)
                }
            case "POST":
                switch path {
                case "home":
    				client.name = params["username"]
    				client.serverWriter.WriteString(client.name+"\n")
    				client.serverWriter.Flush()
                    client.sendQuestion(version)
				case "option":
					fmt.Println(params["option"])
    				client.serverWriter.WriteString(params["option"]+"\n")
    				client.serverWriter.Flush()
    				res, err := client.serverReader.ReadString('\n')
    				res = strip(res)
    				checkErr(err)
					switch params["option"] {
						case "1":
							client.sendListFriends(res, version)
						case "2":
							client.sendAddFriend(res, version)
						case "3":
							client.sendDeleteFriend(res, version)
						case "4":
							client.sendChat(res, version)
                        case "exit":    
                            client.exit()
					}
                case "choose":
                    client.serverWriter.WriteString(params["choose"]+"\n")
    				client.serverWriter.Flush()
                    res, err := client.serverReader.ReadString('\n')
                    res = strip(res)
                    checkErr(err)
                    client.sendOKHtml(res, version)
                case "chatChoose":
                    client.serverWriter.WriteString(params["chatChoose"]+"\n")
    				client.serverWriter.Flush()
                    for {
                        res, err := client.serverReader.ReadString('\n')
                        res = strip(res)
                        checkErr(err)
                        if(res != "START") {
                            s += res+"<br>"
                        }else{
                            break
                        }
                    }
                    client.sendChatHtml(s, version)
                    go client.webRead(version)
                case "chat":
                    if(params["message"] == ""){
                        if(params["file"] == "") {
                            client.sendChatHtml(s, version)
                        }else{
                            client.serverWriter.WriteString("<img src=\"http://" + IP + "/files/" + params["file"] + "\" width=\"20\" height=\"20\">"+"\n")
    				        client.serverWriter.Flush()
                            s += client.name + ": " + "<img src=\"http://" + IP + "/files/" + params["file"] + "\" width=\"20\" height=\"20\">"
                            client.sendChatHtml(s, version)
                            err := os.WriteFile("./files/"+params["file"], data, 0644)
                            checkErr(err)
                        }
                        break
                    }
                    client.serverWriter.WriteString(params["message"]+"\n")
    				client.serverWriter.Flush()
                    s += client.name+ ": " + params["message"]+"<br>"
                    if(params["message"] == "exit"){
                        s = ""
                        client.sendQuestion(version)
                    }else{
                        if(params["file"] == ""){
                            client.sendChatHtml(s, version)
                        }else{
                            client.serverWriter.WriteString("<img src=\"http://" + IP + "/files/" + params["file"] + "\" width=\"20\" height=\"20\">"+"\n")
    				        client.serverWriter.Flush()
                            s += client.name + ": " + "<img src=\"http://" + IP + "/files/" + params["file"] + "\" width=\"20\" height=\"20\">"
                            client.sendChatHtml(s, version)
                            err := os.WriteFile("./files/"+params["file"], data, 0644)
                            fmt.Println("./files/"+params["file"])
                            checkErr(err)
                        }
                    }
                    
                }
        }
        client.clientConn.Close()
        // client.clientWriter = nil
        // client.clientReader = nil
        fmt.Println("==============")
        client.reconnect()
    }
}

func NewClient(connection net.Conn) *Client {
    serverConn, err := net.Dial("tcp", "140.112.30.32:8888")
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

func web(IP string, port string) {
    listener, _ = net.Listen("tcp", IP+":"+port)
    conn, err := listener.Accept()
    checkErr(err)
    fmt.Println("Client from: ", conn.RemoteAddr())
    client := NewClient(conn)
    client.webWorkFlow(listener)
}

func main() {
    mode := os.Args[1]
    if(len(os.Args) > 3){
        IP = os.Args[2]
        port = os.Args[3]
    }
    fmt.Println(mode)
    question = "Home\n (1) List all friends\n (2) Add friend\n (3) Delete friend\n (4) Choose a chat room"
    switch mode {
    case "console":
        console()
    case "web":
        web(IP, port)
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
