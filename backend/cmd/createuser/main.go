// Membuat akun penjual. Tidak ada pendaftaran mandiri di website.
//
//	go run ./cmd/createuser -email penjual@contoh.com -nama "Nama" -password "kata-sandi-kuat"
//
// Koneksi database diambil dari file config yaml (atau DB_DSN bila diisi).
package main

import (
	"database/sql"
	"flag"
	"log"

	"bakery-scheduler/internal/config"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	email := flag.String("email", "", "email login")
	nama := flag.String("nama", "", "nama penjual")
	pass := flag.String("password", "", "kata sandi (minimal 8 karakter)")
	flag.Parse()
	if *email == "" || *nama == "" || len(*pass) < 8 {
		log.Fatal("email, nama, dan password (minimal 8 karakter) wajib diisi")
	}
	if err := config.Init(); err != nil {
		log.Fatalf("[CONFIG] %v", err)
	}
	db, err := sql.Open("mysql", config.Get().Database.Bakery)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	hash, err := bcrypt.GenerateFromPassword([]byte(*pass), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO users (nama, email, password_hash) VALUES (?,?,?)`, *nama, *email, string(hash)); err != nil {
		log.Fatal(err)
	}
	log.Println("akun dibuat:", *email)
}
