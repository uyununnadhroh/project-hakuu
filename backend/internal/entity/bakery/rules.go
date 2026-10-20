package bakery

import "time"

// Weights bobot AHP: J = jam kirim, Q = jumlah pesanan, W = lama menunggu.
type Weights struct{ J, Q, W float64 }

// Settings parameter produksi (dibaca dari tabel settings + ahp_weights).
type Settings struct {
	CapacityPerDay      int     // kapasitas efektif per hari (pcs)
	OvertimeMax         int     // lembur maksimal per hari (pcs)
	WindowHours         float64 // jendela perencanaan (jam)
	MorningCutoffHour   int     // batas H-1: jam kirim sebelum jam ini wajib diproduksi H-1 (default 11.00)
	MaxWaitHours        float64 // batas normalisasi lama menunggu (aging)
	RoundHours          int     // pembulatan jam alternatif (ke atas)
	MinScoreGap         float64 // selisih skor minimum untuk boleh menggeser order
	MaxOffers           int     // maksimal opsi negosiasi
	HoldMinutes         int     // lama tahanan kapasitas sementara
	ProductionStartHour int     // jam produksi harian dimulai (default 06.00)
	ProductionEndHour   int     // jam produksi harian selesai (default 18.00, juga batas reservasi hari itu)
	LastDeliveryHour    int     // jam kirim paling akhir dalam sehari (default 20.00)
	BigOrderQty         int     // pesanan di atas jumlah ini = pesanan besar (default 400, 0 = nonaktif)
	BigOrderLeadDays    int     // pesanan besar wajib dipesan paling lambat H-n (default 7)
	WarnHours           int     // peringatan At risk: jam sebelum produksi harian selesai
	Weights             Weights
}

type ScoreInput struct {
	Qty       int
	DeliverAt time.Time
	OrderedAt time.Time
}

type Score struct{ J, Q, W, Total float64 }

// Alt jam kirim alternatif hasil negosiasi.
type Alt struct {
	DeliverAt     time.Time
	ProductionDay time.Time
}
