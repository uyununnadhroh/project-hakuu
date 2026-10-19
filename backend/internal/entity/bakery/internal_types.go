package bakery

import (
	"database/sql"
	"time"
)

// Struct di file ini hanya dipakai antara service dan data (tidak dikirim ke frontend).

// DayState kondisi kapasitas satu hari produksi.
type DayState struct {
	Day         time.Time
	Base        int
	OvertimeMax int
	Used        int
	Locked      bool
}

// Free sisa kapasitas normal (tanpa lembur).
func (d DayState) Free() int { return max(0, d.Base-d.Used) }

// OrderInsert data untuk INSERT orders.
type OrderInsert struct {
	CustomerID   *int64
	Jumlah       int
	QtyBox       int
	TotalHarga   float64
	WaktuPesan   time.Time
	WaktuKirim   time.Time
	HariDeadline string
	Catatan      string
	CreatedBy    int64 // 0 disimpan sebagai NULL
}

// ProductPrice harga, status aktif, dan bobot produk saat order dibuat.
type ProductPrice struct {
	Harga float64
	Aktif bool
	Bobot float64
}

// Candidate reservasi order lain yang mungkin digeser.
type Candidate struct {
	ResID, OrderID     int64
	Kode               string
	Qty                int
	DeliverAt, Ordered time.Time
	DeadlineStr        string
	Deadline           time.Time
	Score              float64
}

// OfferRow tawaran negosiasi + data order-nya.
type OfferRow struct {
	ID, OrderID int64
	Status      string
	DeliverAt   time.Time
	ProdDayStr  string
	ProdDay     time.Time
	ResID       sql.NullInt64
	Due         sql.NullTime
	Qty         int
	OrderStatus string
	Kode        string
	OrderedAt   time.Time
}

// ExpiredOffer tawaran yang batas responsnya sudah lewat.
type ExpiredOffer struct {
	ID, OrderID, ResID int64
}

// ScoreRow order yang skornya dihitung ulang.
type ScoreRow struct {
	ID    int64
	Input ScoreInput
}

// RebalanceItem order terjadwal yang mungkin dimajukan.
type RebalanceItem struct {
	ResID, OrderID int64
	Kode           string
	Qty            int
	Deliver, Order time.Time
	DeadlineStr    string
	CurStr         string
	Deadline, Cur  time.Time
	Score          float64
}

// RiskOrder order yang berisiko terlambat.
type RiskOrder struct {
	ID   int64
	Kode string
}

// SchedulingCounts hitungan aksi audit untuk laporan.
type AuditCount struct {
	Aksi  string
	Total int
}
