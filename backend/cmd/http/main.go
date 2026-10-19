/*
bakery-scheduler - backend penjadwalan produksi
- struktur mengikuti pola panakea-be (boot, data, service, delivery, entity)
*/
package main

import (
	"log"

	"bakery-scheduler/internal/boot"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	if err := boot.HTTP(); err != nil {
		log.Println("[HTTP] failed to boot http server due to " + err.Error())
	}
}
