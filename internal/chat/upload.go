package chat

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"libretalk/internal/auth"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// uploadHandler handles multipart uploads under the field name "file"
// and writes metadata into "files" table in db
func UploadHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1) Limit size (optional)
		r.Body = http.MaxBytesReader(w, r.Body, 50<<20) // 50 MiB max

		// authenticate user

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization", http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			http.Error(w, "bad authorization scheme", http.StatusUnauthorized)
			return
		}

		parts := strings.Fields(authHeader)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, "bad authorization header", http.StatusUnauthorized)
			return
		}
		tokenStr := parts[1]

		username, err := auth.ParseToken(tokenStr)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// 2) Parse the multipart form
		if err := r.ParseMultipartForm(50 << 20); err != nil {
			http.Error(w, "file too large", http.StatusBadRequest)
			return
		}

		// 3) Retrieve file and header
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "bad upload", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// 4) Write to disk
		id := uuid.New().String()
		ext := filepath.Ext(header.Filename)
		savedName := id + ext
		uploadDir := "./uploads"
		outPath := filepath.Join(uploadDir, savedName)

		// ensure uploadDir exists
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			fmt.Println("mkdir error:", err)
			return
		}

		out, err := os.Create(outPath)
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			fmt.Println("create file error:", err)
			return
		}
		defer out.Close()

		size, err := io.Copy(out, file)
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			fmt.Println("copy file error:", err)
			return
		}

		// 5) Insert metadata into DB

		_, err = db.Exec(`
            INSERT INTO files (id, uploader, original_name, content_type, size_bytes)
            VALUES (?, ?, ?, ?, ?)`,
			id, username, header.Filename, header.Header.Get("Content-Type"), size,
		)

		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			fmt.Println("db insert error:", err)
			return
		}

		// 6) Return JSON with the file URL
		resp := map[string]interface{}{
			"id":          id,
			"url":         "/uploads/" + savedName,
			"filename":    header.Filename,
			"contentType": header.Header.Get("Content-Type"),
			"sizeBytes":   size,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
