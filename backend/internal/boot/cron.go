package boot

import (
	"context"
	"log"
	"time"

	bakeryService "bakery-scheduler/internal/service/bakery"
)

// runHourly menjalankan job penjadwalan sekali saat start, lalu tiap jam.
func runHourly(svc bakeryService.Service) {
	run := func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		if err := svc.RunHourly(ctx); err != nil {
			log.Printf("[CRON] job per jam gagal: %v", err)
		}
	}
	run()
	for range time.Tick(time.Hour) {
		run()
	}
}
