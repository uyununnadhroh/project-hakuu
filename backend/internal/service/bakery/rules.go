package bakery

// Aturan bisnis murni (tanpa database) sehingga mudah diuji.

import (
	"errors"
	"fmt"
	"math"
	"time"

	"bakery-scheduler/internal/entity/bakery"
)

func DefaultSettings() bakery.Settings {
	return bakery.Settings{
		CapacityPerDay: 800, OvertimeMax: 80, WindowHours: 72,
		MorningCutoffHour: 12, MaxWaitHours: 96, RoundHours: 1,
		MinScoreGap: 0.05, MaxOffers: 3, HoldMinutes: 60,
		ProductionEndHour: 10, PackingMinutes: 30, TravelMinutes: 30,
		LastDeliveryHour: 19, WarnHours: 2,
		Weights: bakery.Weights{J: 0.648, Q: 0.230, W: 0.122},
	}
}

func DayOf(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func AddDays(day time.Time, n int) time.Time { return day.AddDate(0, 0, n) }

// ReadyOffset jam paling awal pesanan bisa dikirim bila diproduksi di hari kirim:
// produksi selesai + packing + perjalanan (default 10.00 + 30 + 30 menit = 11.00).
func ReadyOffset(s bakery.Settings) time.Duration {
	return time.Duration(s.ProductionEndHour)*time.Hour +
		time.Duration(s.PackingMinutes+s.TravelMinutes)*time.Minute
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

// AlternativeTimes mencari rekomendasi jam kirim di TANGGAL YANG SAMA dengan
// permintaan awal (tidak pernah pindah tanggal). Hanya jam yang lebih lambat,
// paling awal jam siap (11.00) dan paling akhir LastDeliveryHour (19.00),
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
