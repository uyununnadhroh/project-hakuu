package bakery

import (
	"errors"
	"math"
	"testing"
	"time"

	"bakery-scheduler/internal/entity/bakery"
)

// at() dan wib ada di integration_test.go (satu package)
func near(a, b, tol float64) bool { return math.Abs(a-b) <= tol }

func TestDeadlineDay(t *testing.T) {
	s := DefaultSettings()
	cases := []struct {
		deliver time.Time
		want    time.Time
	}{
		{at(13, 7, 0), at(12, 0, 0)},   // sebelum jam siap 11.00 -> hanya H-1
		{at(13, 10, 59), at(12, 0, 0)}, // masih H-1
		{at(13, 11, 0), at(13, 0, 0)},  // tepat 11.00 -> boleh hari H
		{at(13, 15, 0), at(13, 0, 0)},
	}
	for _, c := range cases {
		if got := DeadlineDay(c.deliver, s); !got.Equal(c.want) {
			t.Errorf("deadline %v = %v, want %v", c.deliver, got, c.want)
		}
	}
}

func TestEarliestDay(t *testing.T) {
	s := DefaultSettings()
	// kirim Selasa, dicek Senin pagi -> H-1 = Senin
	if got := EarliestDay(at(13, 7, 0), at(12, 8, 0), s); !got.Equal(at(12, 0, 0)) {
		t.Errorf("senin pagi: %v", got)
	}
	// sebelum jam produksi selesai (18.00), hari ini masih tersedia
	if got := EarliestDay(at(13, 7, 0), at(12, 17, 0), s); !got.Equal(at(12, 0, 0)) {
		t.Errorf("senin sore: %v", got)
	}
	// setelah jam produksi selesai, hari ini tidak tersedia
	if got := EarliestDay(at(13, 7, 0), at(12, 18, 0), s); !got.Equal(at(13, 0, 0)) {
		t.Errorf("senin malam: %v", got)
	}
	// jauh hari: paling awal H-1
	if got := EarliestDay(at(20, 15, 0), at(12, 8, 0), s); !got.Equal(at(19, 0, 0)) {
		t.Errorf("jauh hari: %v", got)
	}
	// tidak ada lagi H-2: kirim Rabu pagi -> paling awal Selasa, bukan Senin
	if got := EarliestDay(at(14, 8, 0), at(12, 8, 0), s); !got.Equal(at(13, 0, 0)) {
		t.Errorf("tanpa H-2: %v", got)
	}
}

func TestScore(t *testing.T) {
	s := DefaultSettings()
	now := at(12, 8, 0)
	sc := ComputeScore(bakery.ScoreInput{Qty: 250, DeliverAt: now.Add(23 * time.Hour), OrderedAt: now.Add(-48 * time.Hour)}, now, s)
	if !near(sc.J, 1-23.0/72, 1e-9) || !near(sc.Q, 250.0/800, 1e-9) || !near(sc.W, 0.5, 1e-9) {
		t.Fatalf("komponen salah: %+v", sc)
	}
	if !near(sc.Total, 0.5739, 0.001) {
		t.Errorf("total = %v", sc.Total)
	}
	// jam kirim sudah lewat -> J maksimal; order baru -> W = 0
	late := ComputeScore(bakery.ScoreInput{Qty: 10, DeliverAt: now.Add(-time.Hour), OrderedAt: now}, now, s)
	if late.J != 1 || late.W != 0 {
		t.Errorf("clamp salah: %+v", late)
	}
}

func TestRoundUpToStep(t *testing.T) {
	if got := RoundUpToStep(at(13, 11, 20), 1); !got.Equal(at(13, 12, 0)) {
		t.Errorf("11.20 -> %v", got)
	}
	if got := RoundUpToStep(at(13, 12, 0), 1); !got.Equal(at(13, 12, 0)) {
		t.Errorf("12.00 tepat -> %v", got)
	}
	if got := RoundUpToStep(at(13, 12, 5), 1); !got.Equal(at(13, 13, 0)) {
		t.Errorf("12.05 -> %v", got)
	}
}

func TestAHP(t *testing.T) {
	m := [][]float64{{1, 3, 5}, {1.0 / 3, 1, 2}, {0.2, 0.5, 1}}
	w, cr, err := AHP(m)
	if err != nil {
		t.Fatal(err)
	}
	if !near(w[0], 0.648, 0.002) || !near(w[1], 0.230, 0.002) || !near(w[2], 0.122, 0.002) {
		t.Errorf("bobot = %v", w)
	}
	if cr >= 0.1 {
		t.Errorf("CR = %v, harusnya konsisten", cr)
	}
	bad := [][]float64{{1, 9, 1.0 / 9}, {1.0 / 9, 1, 9}, {9, 1.0 / 9, 1}}
	if _, cr, _ := AHP(bad); cr < 0.1 {
		t.Errorf("matriks kontradiktif harusnya CR >= 0.1, dapat %v", cr)
	}
	if _, _, err := AHP([][]float64{{1, 3}, {1, 1}}); err == nil {
		t.Error("matriks tidak berkebalikan harus ditolak")
	}
}

func TestAlternativeTimes(t *testing.T) {
	s := DefaultSettings()
	now := at(12, 9, 0) // Senin 09.00
	// Senin penuh, Selasa masih sisa 400
	free := func(d time.Time) int {
		if d.Equal(at(12, 0, 0)) {
			return 0
		}
		return 400
	}
	got := AlternativeTimes(at(13, 8, 0), 250, now, s, free)
	if len(got) != 3 {
		t.Fatalf("harusnya 3 opsi, dapat %d: %+v", len(got), got)
	}
	// H-1 (Senin) penuh -> ditawarkan mulai 11.00 dengan produksi hari H (Selasa)
	for i, h := range []int{11, 12, 13} {
		if !got[i].DeliverAt.Equal(at(13, h, 0)) || !got[i].ProductionDay.Equal(at(13, 0, 0)) {
			t.Errorf("opsi %d = %+v", i, got[i])
		}
	}
	// rekomendasi berhenti di 20.00 (tidak ada 21.00), tetap di tanggal yang sama
	late := AlternativeTimes(at(13, 18, 20), 250, now, s, free)
	if len(late) != 2 || !late[0].DeliverAt.Equal(at(13, 19, 0)) || !late[1].DeliverAt.Equal(at(13, 20, 0)) {
		t.Errorf("harusnya hanya 19.00 dan 20.00: %+v", late)
	}
	// rekomendasi tidak pernah sebelum 11.00 walau H-1 masih ada sisa
	for _, a := range AlternativeTimes(at(13, 6, 0), 100, at(12, 8, 0), s, func(time.Time) int { return 800 }) {
		if a.DeliverAt.Before(at(13, 11, 0)) || !DayOf(a.DeliverAt).Equal(at(13, 0, 0)) {
			t.Errorf("rekomendasi %v di luar 11.00-20.00 hari yang sama", a.DeliverAt)
		}
	}
	// permintaan siang yang hari produksinya sudah penuh: jam lain tidak menolong
	none := AlternativeTimes(at(13, 14, 0), 250, now, s, func(time.Time) int { return 0 })
	if len(none) != 0 {
		t.Errorf("harusnya tidak ada opsi: %+v", none)
	}
	// batas H-1 = 11.00, jadi tidak ada opsi sebelum itu di hari produksi sama
	for _, a := range AlternativeTimes(at(13, 7, 0), 100, at(13, 5, 0), s, func(time.Time) int { return 800 }) {
		if a.ProductionDay.Equal(DayOf(a.DeliverAt)) && a.DeliverAt.Hour() < 11 {
			t.Errorf("jam %v terlalu dini untuk produksi hari yang sama", a.DeliverAt)
		}
	}
}

func TestCheckDeliveryHour(t *testing.T) {
	s := DefaultSettings()
	for _, ok := range []time.Time{at(13, 7, 0), at(13, 19, 59), at(13, 20, 0)} {
		if err := CheckDeliveryHour(ok, s); err != nil {
			t.Errorf("%v harusnya boleh: %v", ok, err)
		}
	}
	for _, bad := range []time.Time{at(13, 20, 1), at(13, 22, 0)} {
		if err := CheckDeliveryHour(bad, s); !errors.Is(err, bakery.ErrValidation) {
			t.Errorf("%v harusnya ditolak (lewat 20.00): %v", bad, err)
		}
	}
}

func TestIsRushBigOrder(t *testing.T) {
	s := DefaultSettings()
	now := at(12, 9, 0) // Senin 12 Okt
	cases := []struct {
		name    string
		qty     int
		deliver time.Time
		want    bool
	}{
		{"tepat 400, besok", 400, at(13, 10, 0), false}, // tepat 400 belum termasuk besar
		{"besar, besok", 800, at(13, 10, 0), true},      // kasus dosen
		{"besar, H-6", 401, at(18, 10, 0), true},
		{"besar, tepat H-7", 800, at(19, 7, 0), false},
		{"besar, lebih dari H-7", 800, at(25, 10, 0), false},
	}
	for _, c := range cases {
		if got := IsRushBigOrder(c.qty, c.deliver, now, s); got != c.want {
			t.Errorf("%s: %v, want %v", c.name, got, c.want)
		}
	}
	s.BigOrderQty = 0
	if IsRushBigOrder(800, at(13, 10, 0), now, s) {
		t.Error("batas 0 = aturan nonaktif")
	}
}
