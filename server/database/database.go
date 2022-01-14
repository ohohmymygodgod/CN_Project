package database

import (
    "database/sql"
	// "strings"
	"fmt"

    _ "github.com/mattn/go-sqlite3"
)

func OpenDatabase() *sql.DB {
	db, err := sql.Open("sqlite3", "./database/database.db")
    if err != nil {
        fmt.Println(err)
    }

    sql_table := `
    CREATE TABLE IF NOT EXISTS users(
        username VARCHAR(64) PRIMARY KEY
    );
    `
    db.Exec(sql_table)

	sql_table = `
    CREATE TABLE IF NOT EXISTS relations(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username1 VARCHAR(64),
		username2 VARCHAR(64)
    );
    `
    db.Exec(sql_table)


	sql_table = `
    CREATE TABLE IF NOT EXISTS messages(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user VARCHAR(64),
		rid INTEGER,
		message TEXT NULL
    );
    `
	db.Exec(sql_table)

	return db
}

func ListFriends(db *sql.DB, name string) map[string]bool {
	stmt, err := db.Prepare("SELECT r.username1, r.username2 FROM users as u JOIN relations as r WHERE u.username = ? AND (u.username = r.username1 OR u.username = r.username2)")
	checkErr(err)
	rows, err := stmt.Query(name)
	checkErr(err)
	var username1 string
	var username2 string
	friends := make(map[string]bool)
	defer rows.Close()
	for  rows.Next() {
		err = rows.Scan(&username1, &username2)
		checkErr(err)
		if(name == username1) {
			friends[username2] = true
		}else{
			friends[username1] = true
		}
	}
	return friends
}

func ListNotFriends(db *sql.DB, name string) map[string]bool {
	stmt, err := db.Prepare("SELECT username FROM users")
	checkErr(err)
	rows, err := stmt.Query()
	checkErr(err)
	var username string
	notFriends := make(map[string]bool)
	defer rows.Close()
	for  rows.Next() {
		err = rows.Scan(&username)
		checkErr(err)
		if(name != username) {
			notFriends[username] = true
		}
	}
	friends := ListFriends(db, name)
	for friend := range friends {
		delete(notFriends, friend)
	}
	return notFriends
}


func AddFriend(db *sql.DB, name string, friend string) int64 {
	stmt, err := db.Prepare("INSERT INTO relations (username1, username2) VALUES (?,?)")
	checkErr(err)
	res, err := stmt.Exec(name, friend)
	checkErr(err)
	i, err := res.RowsAffected()
	checkErr(err)
	return i
}

func DeleteFriend(db *sql.DB, name string, friend string) int64 {
	stmt, err := db.Prepare("DELETE FROM relations WHERE (username1 == ? AND username2 == ?) OR (username1 == ? AND username2 == ?)")
	checkErr(err)
	res, err := stmt.Exec(name, friend, friend, name)
	checkErr(err)
	i, err := res.RowsAffected()
	checkErr(err)
	return i
}

func ListHistory(db *sql.DB, rid int64) []string {
	stmt, err := db.Prepare("SELECT id, user, message from messages WHERE rid = ?")
	checkErr(err)
	rows, err := stmt.Query(rid)
	checkErr(err)
	var id int64
	var username string
	var message string
	var ret []string
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&id, &username, &message)
		checkErr(err)
		ret = append(ret, username + ": " + message)
	}
	return ret
}

func NameInDB(db *sql.DB, name string) bool {
	stmt, err := db.Prepare("SELECT username FROM users")
	checkErr(err)
	rows, err := stmt.Query()
	checkErr(err)
	var username string
	var ret bool
	ret = false
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&username)
		checkErr(err)
		if(name == username) {
			ret = true
			break
		}
	}
	return ret
}

func GetRelationID(db *sql.DB, name string, friend string) int64 {
	var rid int64
	err := db.QueryRow("SELECT id FROM relations WHERE (username1 = ? AND username2 = ?) OR (username1 = ? AND username2 = ?)", name, friend, friend, name).Scan(&rid)
	checkErr(err)
	return rid
}

func NewUser(db *sql.DB, name string) {
	stmt, err := db.Prepare("INSERT INTO users(username) values(?)")
	checkErr(err)
	res, err := stmt.Exec(name)
	checkErr(err)
	id, err := res.LastInsertId()
	checkErr(err)
	fmt.Println("id", id)
}

func NewMessage(db *sql.DB, name string, rid int64, message string) {
	stmt, err := db.Prepare("INSERT INTO messages(user, rid, message) values(?,?,?)")
	checkErr(err)
	res, err := stmt.Exec(name, rid, message)
	checkErr(err)
	id, err := res.LastInsertId()
	checkErr(err)
	fmt.Println("rid: ", rid, ", message id: ", id, ", message:", message)
}

func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}