package main

import (
    "database/sql"
    "bufio"
    "fmt"
    "net"
    "strings"
    "strconv"
    "sort"
    "runtime"
    //"reflect"
    "server/database"
    _ "github.com/mattn/go-sqlite3"
)

var allClients map[*Client]int
var db *sql.DB

type Client struct {
    ch         map[*Client]chan string
    name       string
    reader     *bufio.Reader
    writer     *bufio.Writer
    conn       net.Conn
    chat       *Client
}

func (client *Client) ListFriends() []string {
    tmp := database.ListFriends(db, client.name)
    var friends []string
    for friend := range tmp {
        friends = append(friends, friend)
    }
    sort.Strings(friends)
    var ret string
    count := 0
    for _, friend := range friends {
        ret += "(" + strconv.Itoa(count+1) + ")" + " " + friend + " "
        count++
    }
    if(len(friends) == 0){
        ret = "No friends."
    }
    ret += "\n"
    client.writer.WriteString(ret)
    client.writer.Flush()
    return friends
}

func (client *Client) ListNotFriends() []string {
    tmp := database.ListNotFriends(db, client.name)
    var notFriends []string
    for user := range tmp {
        notFriends = append(notFriends, user)
    }
    sort.Strings(notFriends)
    var ret string
    count := 0
    for _, notFriend := range notFriends {
        ret += "(" + strconv.Itoa(count+1) + ")" + " " + notFriend + " "
        count++
    }
    if(len(notFriends) == 0) {
        ret = "No other users."
    }
    ret += "\n"
    client.writer.WriteString(ret)
    client.writer.Flush()
    return notFriends
}

func (client *Client) readOption() int {
    i, err := client.reader.ReadString('\n')
    i = strip(i)
    client.handleErr(err)
    index, err := strconv.Atoi(i)
    client.handleErr(err)
    return index
}

func (client *Client) AddFriend() {
    users := client.ListNotFriends()
    if(len(users) == 0) {
        return
    }
    index := client.readOption()
    if(database.AddFriend(db, client.name, users[index-1]) == int64(1)){
        client.writer.WriteString("Success\n")
        client.writer.Flush()
    }

}

func (client *Client) DeleteFriend() {
    users := client.ListFriends()
    if(len(users) == 0){
        return
    }
    index := client.readOption()
    if(database.DeleteFriend(db, client.name, users[index-1]) == int64(1)){
        client.writer.WriteString("Success\n")
        client.writer.Flush()
    }
}

func (client *Client) Read() {
    client.ch[client.chat] = make(chan string)
    for message := range client.ch[client.chat] {
        if(message == "exit"){
            client.chat = nil
            client.ch[client.chat] = nil
            break
        }
        client.writer.WriteString(message + "\n")
        client.writer.Flush()
	}
}

func (client *Client) Write(rid int64) {
    for {
        message, err := client.reader.ReadString('\n')
        message = strip(message)
        client.handleErr(err)
        if(message == "exit"){
            client.ch[client.chat] <- "exit"
            return
        }
        database.NewMessage(db, client.name, rid, message)
        if(client.chat != nil) {
            if(client.chat.ch[client] != nil) {
                client.chat.ch[client] <- client.name + ": " + message
            }
        }
    }
}

func (client *Client) chatHistory(rid int64) {
    messages := database.ListHistory(db, rid)
    l := len(messages)
    client.writer.WriteString("==========start==========\n")
    for i := 0; i < l; i++ {
        client.writer.WriteString(messages[int64(i)] + "\n")
    }
    client.writer.WriteString("START\n")
    client.writer.Flush()
}

func (client *Client) Chat() {
    users := client.ListFriends()
    if(len(users) == 0){
        return
    }
    index := client.readOption()
    friend := users[index-1]
    for user := range allClients {
        if(user.name == friend){
            client.chat = user
            break
        }
    }
    rid := database.GetRelationID(db, client.name, friend)
    client.chatHistory(rid)
    go client.Write(rid)
    client.Read()
    client.writer.WriteString("FINISH\n")
    client.writer.Flush()
}

func (client *Client) WorkFlow() {
    name, err := client.reader.ReadString('\n')
    name = strip(name)
    client.name = name
	fmt.Println(name)
    client.handleErr(err)
    if(!database.NameInDB(db, client.name)){
        database.NewUser(db, client.name)
    }
    for {
        i, err := client.reader.ReadString('\n')
        i = strip(i)
        client.handleErr(err)
        switch i {
        case "1":
            client.ListFriends()
        case "2":
            client.AddFriend()
        case "3":
            client.DeleteFriend()
        case "4":
            client.Chat()
        case "exit":
            client.exit()
        }
    }
}

func NewClient(connection net.Conn) *Client {
    writer := bufio.NewWriter(connection)
    reader := bufio.NewReader(connection)

    client := &Client{
        conn:     connection,
        reader:   reader,
        writer:   writer,
        ch:       make(map[*Client]chan string),
    }
    go client.WorkFlow()
    return client
}


func main() {
    db = database.OpenDatabase()
    allClients = make(map[*Client]int)
    listener, _ := net.Listen("tcp", "140.112.30.32:8888")
    fmt.Println(listener.Addr())
    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println(err.Error())
        }
        client := NewClient(conn)
        allClients[client] = 1
        fmt.Println("Connected, Current users: ", len(allClients))
    }
}

func (client *Client) exit() {
    client.conn.Close()
    delete(allClients, client)
    client = nil
    runtime.Goexit()
}

func (client *Client) handleErr(err error) {
    if(err != nil) {
        fmt.Println(err.Error())
        client.exit()
    }
}


func strip(s string) string {
	s = strings.Replace(s, "\n", "", -1)
	return s
}
