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
	//reads password without showing it in terminal
	pw, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println()
	//database connection --TODO - change db IP and PORT to silent input like db password for prod
	dsn := fmt.Sprintf("server:%s@tcp(192.168.0.202:3306)/selfchat",
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

	http.HandleFunc("/upload", chat.UploadHandler(database))
	fs := http.FileServer(http.Dir("./uploads"))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", fs))

	addr := ":8081"
	log.Println("Listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
