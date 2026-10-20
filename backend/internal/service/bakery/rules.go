package bakery

// Aturan bisnis murni (tanpa database) sehingga mudah diuji.

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"bakery-scheduler/internal/entity/bakery"
)

func DefaultSettings() bakery.Settings {
	return bakery.Settings{
		CapacityPerDay: 800, OvertimeMax: 80, WindowHours: 72,
		MorningCutoffHour: 11, MaxWaitHours: 96, RoundHours: 1,
		MinScoreGap: 0.05, MaxOffers: 3, HoldMinutes: 60,
		ProductionStartHour: 6, ProductionEndHour: 18,
		LastDeliveryHour: 20, WarnHours: 2,
		Weights: bakery.Weights{J: 0.648, Q: 0.230, W: 0.122},
	}
}

func DayOf(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func AddDays(day time.Time, n int) time.Time { return day.AddDate(0, 0, n) }

// ReadyOffset jam paling awal pesanan boleh diproduksi di hari kirim (batas H-1,
// default 11.00). Jam kirim sebelum jam ini wajib diproduksi H-1.
func ReadyOffset(s bakery.Settings) time.Duration {
	return time.Duration(s.MorningCutoffHour) * time.Hour
}

// ReadyLabel ReadyOffset dalam format jam, mis. "11.00" (untuk pesan ke pengguna).
func ReadyLabel(s bakery.Settings) string {
	m := int(ReadyOffset(s).Minutes())
	return fmt.Sprintf("%02d.%02d", m/60, m%60)
}

// Aturan hari produksi (tidak ada lagi produksi H-2):
//   - utama   : H-1 (sehari sebelum hari kirim)
//   - cadangan: hari H (hari kirim), HANYA bila jam kirim >= jam siap (ReadyOffset, default 11.00)
//
// DeadlineDay = hari produksi paling akhir: hari H bila jam kirim >= jam siap, selain itu H-1.
func DeadlineDay(deliverAt time.Time, s bakery.Settings) time.Time {
	day := DayOf(deliverAt)
	if deliverAt.Before(day.Add(ReadyOffset(s))) {
		return AddDays(day, -1)
	}
	return day
}

// EarliestDay = hari produksi paling awal: H-1, atau lebih lambat bila hari itu
// jam produksinya sudah lewat. Hasil bisa lebih besar dari DeadlineDay, artinya
// produksi H-1 sudah tidak mungkin dan jam kirim perlu dinegosiasikan ke >= jam siap.
func EarliestDay(deliverAt, now time.Time, s bakery.Settings) time.Time {
	today := DayOf(now)
	if now.Hour() >= s.ProductionEndHour {
		today = AddDays(today, 1)
	}
	e := AddDays(DayOf(deliverAt), -1)
	if e.Before(today) {
		e = today
	}
	return e
}

func InWindow(deliverAt, now time.Time, s bakery.Settings) bool {
	return deliverAt.Sub(now).Hours() <= s.WindowHours
}

// ComputeScore: Skor = wJ*J + wQ*Q + wW*W, semua ternormalisasi 0..1.
func ComputeScore(in bakery.ScoreInput, now time.Time, s bakery.Settings) bakery.Score {
	remaining := math.Max(in.DeliverAt.Sub(now).Hours(), 0)
	j := 1 - math.Min(remaining/s.WindowHours, 1)
	q := math.Min(float64(in.Qty)/float64(s.CapacityPerDay), 1)
	w := math.Min(math.Max(now.Sub(in.OrderedAt).Hours(), 0)/s.MaxWaitHours, 1)
	return bakery.Score{J: j, Q: q, W: w, Total: s.Weights.J*j + s.Weights.Q*q + s.Weights.W*w}
}

// RoundUpToStep membulatkan KE ATAS ke kelipatan stepHours sejak tengah malam.
func RoundUpToStep(t time.Time, stepHours int) time.Time {
	if stepHours <= 0 {
		stepHours = 1
	}
	day := DayOf(t)
	step := time.Duration(stepHours) * time.Hour
	k := (t.Sub(day) + step - 1) / step
	return day.Add(k * step)
}

// CheckDeliveryHour menolak jam kirim yang melewati jam kirim terakhir (20.00).
func CheckDeliveryHour(deliverAt time.Time, s bakery.Settings) error {
	if deliverAt.After(DayOf(deliverAt).Add(time.Duration(s.LastDeliveryHour) * time.Hour)) {
		return fmt.Errorf("%w: jam kirim paling lambat %02d.00", bakery.ErrValidation, s.LastDeliveryHour)
	}
	return nil
}

// AlternativeTimes mencari rekomendasi jam kirim di TANGGAL YANG SAMA dengan
// permintaan awal (tidak pernah pindah tanggal). Hanya jam yang lebih lambat,
// paling awal jam siap (11.00) dan paling akhir LastDeliveryHour (20.00),
// dibulatkan ke atas per RoundHours, maksimal MaxOffers pilihan terdekat. Sebuah jam hanya valid bila
// ada hari produksi (H-1, atau hari H bila jam >= jam siap) dengan sisa
// kapasitas cukup. Jadi bila H-1 penuh, yang ditawarkan adalah jam mulai
// 11.00 di hari kirim dengan produksi dialihkan ke hari H.
func AlternativeTimes(req time.Time, qty int, now time.Time, s bakery.Settings, free func(day time.Time) int) []bakery.Alt {
	step := time.Duration(max(s.RoundHours, 1)) * time.Hour
	closing := DayOf(req).Add(time.Duration(s.LastDeliveryHour) * time.Hour)
	start := RoundUpToStep(req, s.RoundHours)
	if !start.After(req) {
		start = start.Add(step)
	}
	ready := ReadyOffset(s)
	// rekomendasi tidak pernah lebih pagi dari jam siap (11.00)
	if earliestRec := DayOf(req).Add(ready); start.Before(earliestRec) {
		start = earliestRec
	}

	var out []bakery.Alt
	for t := start; !t.After(closing) && len(out) < s.MaxOffers; t = t.Add(step) {
		dl := DeadlineDay(t, s)
		for d := EarliestDay(t, now, s); !d.After(dl); d = AddDays(d, 1) {
			if free(d) < qty {
				continue
			}
			if d.Equal(DayOf(t)) && t.Before(d.Add(ready)) {
				continue
			}
			out = append(out, bakery.Alt{DeliverAt: t, ProductionDay: d})
			break
		}
	}
	return out
}

var randomIndex = map[int]float64{1: 0, 2: 0, 3: 0.58, 4: 0.90, 5: 1.12, 6: 1.24, 7: 1.32, 8: 1.41, 9: 1.45, 10: 1.49}

// AHP menghitung bobot dari matriks perbandingan berpasangan (skala Saaty)
// beserta Consistency Ratio. Bobot layak dipakai bila CR < 0,1.
func AHP(m [][]float64) (weights []float64, cr float64, err error) {
	n := len(m)
	if n < 2 || n > 10 {
		return nil, 0, errors.New("jumlah kriteria harus 2-10")
	}
	for i := range m {
		if len(m[i]) != n {
			return nil, 0, errors.New("matriks harus persegi")
		}
		if math.Abs(m[i][i]-1) > 1e-9 {
			return nil, 0, errors.New("diagonal matriks harus 1")
		}
		for j := range m[i] {
			if m[i][j] <= 0 {
				return nil, 0, errors.New("nilai matriks harus positif")
			}
			if math.Abs(m[i][j]*m[j][i]-1) > 1e-3 {
				return nil, 0, fmt.Errorf("elemen [%d][%d] dan [%d][%d] harus saling berkebalikan", i, j, j, i)
			}
		}
	}
	colSum := make([]float64, n)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			colSum[j] += m[i][j]
		}
	}
	weights = make([]float64, n)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			weights[i] += m[i][j] / colSum[j]
		}
		weights[i] /= float64(n)
	}
	var lambda float64
	for i := 0; i < n; i++ {
		var row float64
		for j := 0; j < n; j++ {
			row += m[i][j] * weights[j]
		}
		lambda += row / weights[i]
	}
	lambda /= float64(n)
	ci := math.Max((lambda-float64(n))/float64(n-1), 0)
	if ri := randomIndex[n]; ri > 0 {
		cr = ci / ri
	}
	return weights, cr, nil
}

// CheckVariants memvalidasi pilihan varian satu baris order.
//   - maks = 0 (produk biasa): tidak boleh ada varian.
//   - maks > 0 (paket pilih varian): tepat maks varian, tanpa duplikat, semuanya
//     milik paket itu; varian yang punya opsi tambahan wajib memilih salah satu
//     opsinya, varian tanpa opsi tidak boleh membawa opsi.
func CheckVariants(maks int, avail []bakery.ProductVariant, chosen []bakery.ItemVariant) error {
	if maks <= 0 {
		if len(chosen) > 0 {
			return errors.New("produk ini tidak memakai varian")
		}
		return nil
	}
	if len(chosen) != maks {
		return fmt.Errorf("pilih tepat %d varian (dipilih %d)", maks, len(chosen))
	}
	byID := make(map[int64]bakery.ProductVariant, len(avail))
	for _, v := range avail {
		byID[v.ID] = v
	}
	seen := make(map[int64]bool, len(chosen))
	for _, c := range chosen {
		v, ok := byID[c.VariantID]
		if !ok {
			return fmt.Errorf("varian %d tidak tersedia untuk paket ini", c.VariantID)
		}
		if seen[c.VariantID] {
			return fmt.Errorf("varian %s dipilih lebih dari sekali", v.Nama)
		}
		seen[c.VariantID] = true
		if len(v.OpsiPilihan) == 0 {
			if c.Opsi != "" {
				return fmt.Errorf("varian %s tidak punya pilihan tambahan", v.Nama)
			}
			continue
		}
		valid := false
		for _, o := range v.OpsiPilihan {
			if o == c.Opsi {
				valid = true
				break
			}
		}
		if !valid {
			label := v.OpsiLabel
			if label == "" {
				label = "pilihan"
			}
			return fmt.Errorf("pilih %s yang valid untuk varian %s", label, v.Nama)
		}
	}
	return nil
}

// Batas paket pilih varian.
const (
	MaxIsiPaket   = 10 // isi paket paling banyak
	MaxOpsiVarian = 20 // pilihan tambahan per varian paling banyak
)

// NormVariants memeriksa dan merapikan daftar varian sebuah produk.
//   - maks 0 = produk biasa: tidak boleh punya varian
//   - nama varian wajib, maksimal 60 karakter, tidak boleh dobel (tanpa beda huruf besar/kecil)
//   - pilihan tambahan (mis. isian selai) tidak boleh kosong/dobel; labelnya bawaan "Pilihan"
//   - varian aktif minimal sebanyak isi paket, supaya pesanan selalu bisa dipenuhi
func NormVariants(maks int, in []bakery.VariantInput) ([]bakery.ProductVariant, error) {
	if maks < 0 || maks > MaxIsiPaket {
		return nil, fmt.Errorf("%w: isi paket harus 0 sampai %d varian", bakery.ErrValidation, MaxIsiPaket)
	}
	if maks == 0 && len(in) > 0 {
		return nil, fmt.Errorf("%w: varian hanya untuk paket pilih varian (isi paket lebih dari 0)", bakery.ErrValidation)
	}
	out := make([]bakery.ProductVariant, 0, len(in))
	seen := map[string]bool{}
	aktif := 0
	for _, v := range in {
		nama := strings.TrimSpace(v.Nama)
		if nama == "" || len([]rune(nama)) > 60 {
			return nil, fmt.Errorf("%w: nama varian wajib diisi (maksimal 60 karakter)", bakery.ErrValidation)
		}
		key := strings.ToLower(nama)
		if seen[key] {
			return nil, fmt.Errorf("%w: varian %s ditulis lebih dari sekali", bakery.ErrValidation, nama)
		}
		seen[key] = true

		var opsi []string
		seenOpsi := map[string]bool{}
		for _, o := range v.OpsiPilihan {
			o = strings.TrimSpace(o)
			if o == "" || len([]rune(o)) > 60 {
				return nil, fmt.Errorf("%w: pilihan untuk varian %s tidak boleh kosong (maksimal 60 karakter)", bakery.ErrValidation, nama)
			}
			if seenOpsi[strings.ToLower(o)] {
				return nil, fmt.Errorf("%w: pilihan %s pada varian %s ditulis lebih dari sekali", bakery.ErrValidation, o, nama)
			}
			seenOpsi[strings.ToLower(o)] = true
			opsi = append(opsi, o)
		}
		if len(opsi) > MaxOpsiVarian {
			return nil, fmt.Errorf("%w: varian %s maksimal %d pilihan", bakery.ErrValidation, nama, MaxOpsiVarian)
		}
		label := ""
		if len(opsi) > 0 {
			label = strings.TrimSpace(v.OpsiLabel)
			if label == "" {
				label = "Pilihan"
			}
			if len([]rune(label)) > 60 {
				return nil, fmt.Errorf("%w: label pilihan varian %s maksimal 60 karakter", bakery.ErrValidation, nama)
			}
		}
		on := v.Aktif == nil || *v.Aktif
		if on {
			aktif++
		}
		out = append(out, bakery.ProductVariant{ID: v.ID, Nama: nama, OpsiLabel: label, OpsiPilihan: opsi, Aktif: on})
	}
	if aktif < maks {
		return nil, fmt.Errorf("%w: paket isi %d butuh minimal %d varian aktif (baru %d)", bakery.ErrValidation, maks, maks, aktif)
	}
	return out, nil
}

// ------------------------- KONFIRMASI PENGIRIMAN -------------------------

const (
	JamBatchPagi  = 11 // dibuka 11.00 WIB: pesanan kirim hari ini sebelum 11.00
	JamBatchMalam = 20 // dibuka 20.00 WIB: pesanan kirim hari ini mulai 11.00
)

// WIB zona waktu semua perhitungan konfirmasi pengiriman (Asia/Jakarta, UTC+7,
// tanpa musim panas), terlepas dari zona waktu server atau APP_TZ.
var WIB = time.FixedZone("WIB", 7*60*60)

// BatchKirim batch konfirmasi pengiriman untuk sebuah jam kirim: "pagi" atau "malam".
func BatchKirim(kirim time.Time) string {
	if kirim.In(WIB).Hour() < JamBatchPagi {
		return "pagi"
	}
	return "malam"
}

// BatchDibuka apakah batch sudah boleh dikonfirmasi pada waktu now.
func BatchDibuka(batch string, now time.Time) bool {
	now = now.In(WIB)
	switch batch {
	case "pagi":
		return now.Hour() >= JamBatchPagi
	case "malam":
		return now.Hour() >= JamBatchMalam
	}
	return false
}

// CheckDeliveryConfirm aturan konfirmasi pengiriman satu pesanan:
//   - pilihan hanya "selesai" atau "belum_terkirim"
//   - pesanan selesai tidak bisa diubah lagi; belum_terkirim boleh diubah ke selesai
//   - pesanan batal / belum terjadwal tidak bisa dikonfirmasi
//   - hanya untuk pesanan yang tanggal kirimnya sudah tiba dan batch-nya sudah dibuka
//
// Tanggal & jam dihitung dalam WIB.
func CheckDeliveryConfirm(orderStatus, to string, kirim, now time.Time) error {
	kirim, now = kirim.In(WIB), now.In(WIB)
	if to != "selesai" && to != "belum_terkirim" {
		return fmt.Errorf("%w: status konfirmasi harus selesai atau belum_terkirim", bakery.ErrValidation)
	}
	switch orderStatus {
	case "selesai":
		return fmt.Errorf("%w: pesanan sudah selesai dan tidak bisa diubah lagi", bakery.ErrState)
	case "batal":
		return fmt.Errorf("%w: pesanan sudah dibatalkan", bakery.ErrState)
	case "baru", "menunggu_negosiasi":
		return fmt.Errorf("%w: pesanan belum dijadwalkan", bakery.ErrState)
	}
	today, day := DayOf(now), DayOf(kirim)
	if day.After(today) {
		return fmt.Errorf("%w: pesanan baru dikirim %s", bakery.ErrState, day.Format("2006-01-02"))
	}
	if day.Equal(today) && !BatchDibuka(BatchKirim(kirim), now) {
		jam := JamBatchPagi
		if BatchKirim(kirim) == "malam" {
			jam = JamBatchMalam
		}
		return fmt.Errorf("%w: konfirmasi batch %s dibuka mulai %02d.00", bakery.ErrState, BatchKirim(kirim), jam)
	}
	return nil
}
