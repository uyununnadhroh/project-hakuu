package boot

import (
	"log"
	"net/http"
	"time"

	"bakery-scheduler/internal/config"

	"github.com/jmoiron/sqlx"

	bakeryData "bakery-scheduler/internal/data/bakery"
	bakeryHandler "bakery-scheduler/internal/delivery/http/bakery"
	bakeryService "bakery-scheduler/internal/service/bakery"

	server "bakery-scheduler/internal/delivery/http"
)

// HTTP will load configuration, do dependency injection and then start the HTTP server
func HTTP() error {
	var (
		s   server.Server
		cfg *config.Config // Configuration object
	)

	err := config.Init()
	if err != nil {
		log.Fatalf("[CONFIG] Failed to initialize config: %v", err)
	}
	cfg = config.Get()

	if cfg.Database.Bakery == "" || len(cfg.JWT.Secret) < 32 {
		log.Fatal("[CONFIG] database.bakery wajib diisi dan jwt.secret minimal 32 karakter")
	}

	loc, err := time.LoadLocation(cfg.Server.Timezone)
	if err != nil {
		log.Fatalf("[CONFIG] Failed to load timezone %s: %v", cfg.Server.Timezone, err)
	}

	db, err := openConnectionPool("mysql", cfg.Database.Bakery)
	if err != nil {
		log.Fatalf("[DB] Failed to initialize database connection: %v", err)
	}

	bkd := bakeryData.New(db)
	bks := bakeryService.New(bkd, loc, []byte(cfg.JWT.Secret))
	bkh := bakeryHandler.New(bks)

	// Job per jam (di dalam server): lepas tahanan kedaluwarsa, hitung ulang skor,
	// tandai terlambat, kunci jadwal, majukan order, tandai At risk.
	go runHourly(bks)

	s = server.Server{
		Bakery:         bkh,
		JWTSecret:      []byte(cfg.JWT.Secret),
		AllowedOrigins: cfg.CORS.Origins,
	}

	if err := s.Serve(cfg.Server.Port); err != http.ErrServerClosed {
		return err
	}

	return nil
}

func openConnectionPool(driver string, connString string) (db *sqlx.DB, err error) {
	db, err = sqlx.Open(driver, connString)
	if err != nil {
		return db, err
	}
	db.SetMaxOpenConns(20)
	db.SetConnMaxLifetime(5 * time.Minute)

	err = db.Ping()
	if err != nil {
		log.Printf("[ERROR] %s ping connection failed (cek XAMPP MySQL sudah Start dan DSN di config benar)", driver)
		return db, err
	}

	return db, err
}
