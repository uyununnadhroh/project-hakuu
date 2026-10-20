package bakery

import (
	"encoding/json"
	"time"
)

// ======================= REQUEST (body JSON dari frontend) =======================

type ReqLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ReqCustomer struct {
	Nama    string `json:"nama"`
	Telepon string `json:"telepon"`
	Alamat  string `json:"alamat"`
}

type ReqCreateOrder struct {
	CustomerID *int64 `json:"customer_id"`
	// Customer diisi bila pembeli belum tercatat; pembeli yang sama dipakai ulang.
	Customer  *ReqCustomer `json:"customer"`
	Items     []Item       `json:"items"`
	DeliverAt string       `json:"deliver_at"`
	Note      string       `json:"note"`
	// BahanSiap jawaban penjual atas hasil perlu_cek_bahan: kosong = belum ditanya.
	BahanSiap *bool `json:"bahan_siap"`
}

type ReqRespondOffer struct {
	Accept bool `json:"accept"`
}

type ReqWeights struct {
	Matrix [][]float64 `json:"matrix"`
}

type ReqProduct struct {
	Nama     string   `json:"nama"`
	Harga    float64  `json:"harga"`
	Kategori string   `json:"kategori"`
	Jenis    string   `json:"jenis"`
	Bobot    *float64 `json:"bobot"`
	// Paket pilih varian: isi paket (0 = produk biasa) dan daftar variannya.
	MaksVarian int            `json:"maks_varian"`
	Variants   []VariantInput `json:"variants"`
}

type ReqPatchProduct struct {
	Nama     *string  `json:"nama"`
	Harga    *float64 `json:"harga"`
	Aktif    *bool    `json:"aktif"`
	Kategori *string  `json:"kategori"`
	Jenis    *string  `json:"jenis"`
	Bobot    *float64 `json:"bobot"`
	// Variants bila diisi menggantikan seluruh daftar varian produk:
	// varian lama yang tidak dikirim dihapus (atau dinonaktifkan bila sudah pernah dipesan).
	MaksVarian *int            `json:"maks_varian"`
	Variants   *[]VariantInput `json:"variants"`
}

// VariantInput satu varian paket. ID kosong = varian baru.
type VariantInput struct {
	ID          int64    `json:"id,omitempty"`
	Nama        string   `json:"nama"`
	OpsiLabel   string   `json:"opsi_label"`
	OpsiPilihan []string `json:"opsi_pilihan"`
	Aktif       *bool    `json:"aktif"` // kosong = aktif
}

type ReqCategory struct {
	Nama string `json:"nama"`
}

// ======================= AUTH =======================

type User struct {
	ID           int64
	Nama         string
	Email        string
	PasswordHash string
	Aktif        bool
}

type LoginResult struct {
	Token string `json:"token"`
	Nama  string `json:"nama"`
	Email string `json:"email"`
}

// UserView akun yang bisa masuk ke website (GET /api/users). Password tidak ikut dikirim.
type UserView struct {
	ID        int64  `json:"id"`
	Nama      string `json:"nama"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

// Actor user yang sedang login (diambil dari token JWT).
type Actor struct {
	UserID int64
	Nama   string
}

// ======================= ORDER =======================

type Item struct {
	ProductID int64 `json:"product_id"`
	Qty       int   `json:"qty"`
	// Variants wajib untuk paket pilih varian (jumlahnya = maks_varian produk),
	// berlaku untuk semua box di baris ini. Produk biasa: kosong.
	Variants []ItemVariant `json:"variants,omitempty"`
}

type ItemVariant struct {
	VariantID int64  `json:"variant_id"`
	Opsi      string `json:"opsi,omitempty"` // pilihan tambahan, mis. isian selai untuk Lava
}

type NewOrder struct {
	CustomerID *int64
	Items      []Item
	DeliverAt  time.Time
	Note       string
	Actor      string
	CreatedBy  int64 // id user yang menginput order (0 = tidak diketahui)
	// StockReady jawaban penjual untuk pesanan besar mendadak: nil = tanyakan
	// dulu (hasil perlu_cek_bahan), true = bahan baku cukup, false = tolak.
	StockReady *bool
}

type Offer struct {
	ID            int64  `json:"id"`
	Seq           int    `json:"seq"`
	DeliverAt     string `json:"deliver_at"`
	ProductionDay string `json:"production_day"`
}

type Move struct {
	OrderID int64  `json:"order_id"`
	Kode    string `json:"kode"`
	From    string `json:"from"`
	To      string `json:"to"`
	Qty     int    `json:"qty"`
}

const (
	OutcomeAccepted    = "diterima"
	OutcomeNegotiation = "perlu_negosiasi"
	OutcomeRejected    = "ditolak"
	OutcomeCheckStock  = "perlu_cek_bahan" // pesanan besar mendadak; menunggu jawaban penjual
)

// Result hasil POST /api/orders.
type Result struct {
	Outcome       string   `json:"outcome"`
	OrderID       int64    `json:"order_id,omitempty"`
	Kode          string   `json:"kode,omitempty"`
	ProductionDay string   `json:"production_day,omitempty"`
	Overtime      int      `json:"overtime_pcs,omitempty"`
	Moved         []Move   `json:"moved,omitempty"`
	Offers        []Offer  `json:"offers,omitempty"`
	AltDates      []string `json:"alternative_dates,omitempty"`
	Reason        string   `json:"reason,omitempty"`
}

type OrderRow struct {
	ID            int64    `json:"id"`
	Kode          string   `json:"kode"`
	Customer      *string  `json:"customer,omitempty"`
	Qty           int      `json:"qty"`     // beban kapasitas (box x bobot produk)
	QtyBox        int      `json:"qty_box"` // jumlah box yang dipesan
	Products      string   `json:"products"`
	OrderedAt     string   `json:"ordered_at"`
	DeliverAt     string   `json:"deliver_at"`
	ProductionDay *string  `json:"production_day"`
	Status        string   `json:"status"`
	StatusLabel   string   `json:"status_label"`
	NeedsNego     bool     `json:"needs_negotiation"`
	Score         *float64 `json:"score"`
	Overtime      int      `json:"overtime_pcs"`
	CanStart      bool     `json:"can_start"` // tombol "Mulai produksi" hanya aktif pada/ setelah hari produksi
	// Penanggung jawab: akun yang menginput order (kosong untuk order lama)
	CreatedByNama  string `json:"created_by_nama"`
	CreatedByEmail string `json:"created_by_email"`
}

type OrderItemView struct {
	ID          int64                  `json:"-"`
	Produk      string                 `json:"produk"`
	Qty         int                    `json:"qty"`
	HargaSatuan float64                `json:"harga_satuan"`
	Subtotal    float64                `json:"subtotal"`
	Variants    []OrderItemVariantView `json:"variants,omitempty"`
}

type OrderItemVariantView struct {
	Nama string `json:"nama"`
	Opsi string `json:"opsi,omitempty"`
}

type OrderOfferView struct {
	ID            int64  `json:"id"`
	Seq           int    `json:"seq"`
	DeliverAt     string `json:"deliver_at"`
	ProductionDay string `json:"production_day"`
	Status        string `json:"status"`
	RespondBefore string `json:"respond_before,omitempty"`
}

type OrderHistoryView struct {
	Waktu  string `json:"waktu"`
	Aksi   string `json:"aksi"`
	Alasan string `json:"alasan"`
	Pelaku string `json:"pelaku"`
}

type OrderDetail struct {
	OrderRow
	Phone      string             `json:"phone"`
	Address    string             `json:"address"`
	Note       string             `json:"note"`
	Total      float64            `json:"total_harga"`
	ScoreParts map[string]float64 `json:"score_parts,omitempty"`
	Items      []OrderItemView    `json:"items"`
	Offers     []OrderOfferView   `json:"offers"`
	History    []OrderHistoryView `json:"history"`
}

// OrderExtra kolom tambahan untuk detail order (dibaca data layer).
type OrderExtra struct {
	Total   float64
	J, Q, W *float64
	Phone   string
	Address string
	Note    string
}

type DayView struct {
	Date     string     `json:"date"`
	Capacity int        `json:"capacity"`
	Used     int        `json:"used"`
	Overtime int        `json:"overtime_pcs"`
	Locked   bool       `json:"locked"`
	Orders   []OrderRow `json:"orders"`
}

// ======================= KONFIRMASI PENGIRIMAN =======================

// ReqConfirmDelivery POST /api/delivery-confirmations/{id} (id = order).
type ReqConfirmDelivery struct {
	Status string `json:"status"` // selesai | belum_terkirim
	Alasan string `json:"alasan"` // opsional, hanya untuk belum_terkirim
}

// DeliveryConfirmation konfirmasi terakhir sebuah pesanan.
type DeliveryConfirmation struct {
	Status string    `json:"status"`
	Alasan string    `json:"alasan,omitempty"`
	Waktu  string    `json:"waktu"`
	Oleh   string    `json:"oleh,omitempty"`
	At     time.Time `json:"-"`
}

type DeliveryOrder struct {
	ID         int64                 `json:"id"`
	Kode       string                `json:"kode"`
	Customer   string                `json:"customer"`
	Products   string                `json:"products"`
	DeliverAt  string                `json:"deliver_at"`
	Status     string                `json:"status"`     // status pesanan
	Konfirmasi *DeliveryConfirmation `json:"konfirmasi"` // nil = belum dikonfirmasi
	Kirim      time.Time             `json:"-"`
}

type DeliveryBatch struct {
	Key        string          `json:"key"` // pagi | malam | sebelumnya
	Label      string          `json:"label"`
	Keterangan string          `json:"keterangan"`
	Orders     []DeliveryOrder `json:"orders"`
}

// DeliveryView GET /api/delivery-confirmations: hanya batch yang sudah dibuka.
type DeliveryView struct {
	Date    string          `json:"date"`
	Batches []DeliveryBatch `json:"batches"`
	Pending int             `json:"pending"` // pesanan di batch terbuka yang belum Selesai (notifikasi aktif bila > 0)
}

// ======================= MASTER DATA =======================

type Product struct {
	ID       int64   `json:"id"`
	Nama     string  `json:"nama"`
	Harga    float64 `json:"harga"`
	Aktif    bool    `json:"aktif"`
	Kategori string  `json:"kategori"`
	Jenis    string  `json:"jenis"`
	Bobot    float64 `json:"bobot"`
	// MaksVarian > 0: paket pilih varian, pembeli wajib memilih tepat sejumlah ini.
	MaksVarian int              `json:"maks_varian"`
	Variants   []ProductVariant `json:"variants"`
}

type ProductVariant struct {
	ID          int64    `json:"id"`
	ProductID   int64    `json:"-"`
	Nama        string   `json:"nama"`
	OpsiLabel   string   `json:"opsi_label,omitempty"`   // mis. "Isian selai"
	OpsiPilihan []string `json:"opsi_pilihan,omitempty"` // mis. ["Stroberi","Blueberry"]
	Aktif       bool     `json:"aktif"`
}

type ProductInput struct {
	Nama       string
	Harga      float64
	Kategori   string
	Jenis      string
	Bobot      *float64
	MaksVarian int
	Variants   []VariantInput
}

type Customer struct {
	ID      int64  `json:"id"`
	Nama    string `json:"nama"`
	Telepon string `json:"telepon"`
}

type Notification struct {
	ID    int64  `json:"id"`
	Pesan string `json:"pesan"`
	Waktu string `json:"waktu"`
}

type AuditEntry struct {
	Waktu   string          `json:"waktu"`
	OrderID *int64          `json:"order_id,omitempty"`
	Order   string          `json:"order"`
	Aksi    string          `json:"aksi"`
	Alasan  string          `json:"alasan"`
	Pelaku  string          `json:"pelaku"`
	Sebelum json.RawMessage `json:"sebelum,omitempty"`
	Sesudah json.RawMessage `json:"sesudah,omitempty"`
}

// ======================= LAPORAN =======================

type Summary struct {
	Revenue  float64 `json:"revenue"`
	Orders   int     `json:"orders"`
	Pcs      int     `json:"pcs"`
	AvgOrder float64 `json:"avg_order_value"`
}

type DayRevenue struct {
	Date    string  `json:"date"`
	Orders  int     `json:"orders"`
	Pcs     int     `json:"pcs"`
	Revenue float64 `json:"revenue"`
}

type ProductRow struct {
	Name    string  `json:"name"`
	Pcs     int     `json:"pcs"`
	Revenue float64 `json:"revenue"`
	Share   float64 `json:"share_pct"`
}

type CustomerRow struct {
	Name   string  `json:"name"`
	Orders int     `json:"orders"`
	Spend  float64 `json:"spend"`
}

type Scheduling struct {
	Delivered      int      `json:"delivered_orders"`
	OnTime         int      `json:"on_time_orders"`
	OnTimePct      *float64 `json:"on_time_pct"`
	UtilizationPct *float64 `json:"capacity_utilization_pct"`
	OvertimeDays   int      `json:"overtime_days"`
	OvertimePcs    int      `json:"overtime_pcs"`
	Moved          int      `json:"orders_moved"`
	Advanced       int      `json:"orders_advanced"`
	NegOffered     int      `json:"negotiation_offered"`
	NegAccepted    int      `json:"negotiation_accepted"`
	NegDeclined    int      `json:"negotiation_declined_or_expired"`
	Rejected       int      `json:"orders_rejected"`
}

type Report struct {
	Month            string         `json:"month"`
	Closed           bool           `json:"closed"`
	ClosedAt         string         `json:"closed_at,omitempty"`
	Summary          Summary        `json:"summary"`
	Previous         Summary        `json:"previous_month"`
	RevenueChangePct *float64       `json:"revenue_change_pct"`
	Daily            []DayRevenue   `json:"daily"`
	TopProducts      []ProductRow   `json:"top_products"`
	TopCustomers     []CustomerRow  `json:"top_customers"`
	Status           map[string]int `json:"order_status"`
	Scheduling       Scheduling     `json:"scheduling"`
}
