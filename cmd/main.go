package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"libretalk/internal/chat"
	"libretalk/internal/db"

	"golang.org/x/term"
)

func main() {
	fmt.Print("Enter DB password: ")
	pw, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println()

	dsn := fmt.Sprintf("server:%s@tcp(chatdb.s:3306)/selfchat",
		strings.TrimSpace(string(pw)),
	)
	database, err := db.Connect(dsn)
	if err != nil {
		log.Fatal("DB connection error:", err)
	}
	defer database.Close()
	log.Println("DB connected")

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		chat.Handler(w, r, database)
	})

	addr := ":8081"
	log.Println("Listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
