package bakery

import "time"

// Weights bobot AHP: J = jam kirim, Q = jumlah pesanan, W = lama menunggu.
type Weights struct{ J, Q, W float64 }

// Settings parameter produksi (dibaca dari tabel settings + ahp_weights).
type Settings struct {
	CapacityPerDay    int     // kapasitas efektif per hari (pcs)
	OvertimeMax       int     // lembur maksimal per hari (pcs)
	WindowHours       float64 // jendela perencanaan (jam)
	MorningCutoffHour int     // TIDAK DIPAKAI LAGI: batas hari H kini = jam siap (jam_produksi_selesai + packing + tempuh)
	MaxWaitHours      float64 // batas normalisasi lama menunggu (aging)
	RoundHours        int     // pembulatan jam alternatif (ke atas)
	MinScoreGap       float64 // selisih skor minimum untuk boleh menggeser order
	MaxOffers         int     // maksimal opsi negosiasi
	HoldMinutes       int     // lama tahanan kapasitas sementara
	ProductionEndHour int     // jam produksi harian selesai (juga batas reservasi hari itu)
	PackingMinutes    int
	TravelMinutes     int
	LastDeliveryHour  int // jam kirim paling akhir dalam sehari
	WarnHours         int // peringatan At risk: jam sebelum produksi harian selesai
	Weights           Weights
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
