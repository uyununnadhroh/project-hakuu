package bakery

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	bakeryData "bakery-scheduler/internal/data/bakery"
	"bakery-scheduler/internal/entity/bakery"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// Dijalankan hanya bila TEST_DB_DSN diisi (database KOSONG berisi db/schema.sql).
var wib = time.FixedZone("WIB", 7*3600)

func at(day, h, m int) time.Time { return time.Date(2026, 10, day, h, m, 0, 0, wib) }

type env struct {
	db  *sqlx.DB
	svc *Service
	now *time.Time
	pid int64
}

func setup(t *testing.T, capacity, overtime int) *env {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Skip("TEST_DB_DSN tidak diisi")
	}
	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	for _, q := range []string{
		"DELETE FROM notifications", "DELETE FROM audit_log", "DELETE FROM negotiation_offers",
		"DELETE FROM capacity_reservations", "DELETE FROM order_items", "DELETE FROM orders",
		"DELETE FROM daily_capacity", "DELETE FROM products",
	} {
		if _, err := db.Exec(q); err != nil {
			t.Fatal(err)
		}
	}
	db.Exec(`UPDATE settings SET nilai=? WHERE kunci='kapasitas_harian'`, capacity)
	db.Exec(`UPDATE settings SET nilai=? WHERE kunci='lembur_maks'`, overtime)
	r, err := db.Exec(`INSERT INTO products (nama, harga) VALUES ('Roti', 1000)`)
	if err != nil {
		t.Fatal(err)
	}
	pid, _ := r.LastInsertId()
	now := at(12, 8, 0) // Senin 08.00
	e := &env{now: &now, pid: pid}
	e.db = db
	svc := New(bakeryData.New(db), wib, []byte("rahasia-tes"))
	svc.Now = func() time.Time { return *e.now }
	e.svc = &svc
	return e
}

func (e *env) order(t *testing.T, qty int, deliver time.Time) *bakery.Result {
	t.Helper()
	res, err := e.svc.AcceptOrder(context.Background(), bakery.NewOrder{Items: []bakery.Item{{ProductID: e.pid, Qty: qty}}, DeliverAt: deliver})
	if err != nil {
		t.Fatalf("AcceptOrder: %v", err)
	}
	return res
}

func (e *env) used(t *testing.T, day string) int {
	var n int
	if err := e.db.QueryRow(`SELECT COALESCE(SUM(jumlah),0) FROM capacity_reservations WHERE tanggal=?`, day).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// Skenario pembahasan: A, B, C masuk Senin pagi, lalu order mendadak G
// menggeser C ke Selasa (hari kirim C, jam kirim >= 11.00). Order A sudah
// diproduksi sehingga tidak boleh digeser.
func TestScenarioShiftLowerPriority(t *testing.T) {
	e := setup(t, 400, 40)
	a := e.order(t, 250, at(13, 7, 0))  // kirim Selasa 07.00 -> produksi Senin
	b := e.order(t, 200, at(13, 15, 0)) // kirim Selasa 15.00 -> Senin penuh -> Selasa
	c := e.order(t, 150, at(13, 15, 0)) // kirim Selasa 15.00 -> H-1 Senin masih sisa 150
	if a.ProductionDay != "2026-10-12" || b.ProductionDay != "2026-10-13" || c.ProductionDay != "2026-10-12" {
		t.Fatalf("reservasi awal salah: A=%s B=%s C=%s", a.ProductionDay, b.ProductionDay, c.ProductionDay)
	}
	if err := e.svc.Transition(context.Background(), a.OrderID, "start", "penjual"); err != nil {
		t.Fatal(err)
	}
	*e.now = at(12, 9, 0)
	g := e.order(t, 150, at(13, 8, 0)) // mendadak, harus Senin
	if g.Outcome != bakery.OutcomeAccepted || g.ProductionDay != "2026-10-12" {
		t.Fatalf("G harusnya diterima di Senin: %+v", g)
	}
	if len(g.Moved) != 1 || g.Moved[0].OrderID != c.OrderID || g.Moved[0].To != "2026-10-13" {
		t.Fatalf("yang digeser harusnya hanya C ke Selasa: %+v", g.Moved)
	}
	if e.used(t, "2026-10-12") != 400 || e.used(t, "2026-10-13") != 350 {
		t.Errorf("kapasitas salah: senin=%d selasa=%d", e.used(t, "2026-10-12"), e.used(t, "2026-10-13"))
	}
	var hist int
	e.db.QueryRow(`SELECT COUNT(*) FROM audit_log WHERE order_id=? AND aksi='digeser'`, c.OrderID).Scan(&hist)
	if hist != 1 {
		t.Errorf("audit penggeseran tidak tercatat")
	}
}

// Kapasitas penuh, tidak ada yang bisa digeser, lembur tidak cukup -> sistem
// merekomendasikan jam kirim lebih lambat di hari yang sama.
func TestScenarioNegotiation(t *testing.T) {
	e := setup(t, 400, 40)
	*e.now = at(12, 9, 0)
	e.order(t, 400, at(13, 8, 0)) // memenuhi Senin
	y := e.order(t, 250, at(13, 8, 0))
	if y.Outcome != bakery.OutcomeNegotiation || len(y.Offers) != 3 {
		t.Fatalf("harusnya 3 opsi negosiasi: %+v", y)
	}
	for i, h := range []string{"11:00", "12:00", "13:00"} { // H-1 penuh -> mulai 11.00, produksi hari H
		if y.Offers[i].DeliverAt != "2026-10-13 "+h || y.Offers[i].ProductionDay != "2026-10-13" {
			t.Errorf("opsi %d = %+v", i, y.Offers[i])
		}
	}
	ctx := context.Background()
	due, err := e.svc.OfferToBuyer(ctx, y.Offers[0].ID)
	if err != nil || !due.Equal(at(12, 10, 0)) {
		t.Fatalf("OfferToBuyer: due=%v err=%v", due, err)
	}
	if e.used(t, "2026-10-13") != 250 { // tahanan sementara
		t.Errorf("kapasitas harusnya tertahan 250, dapat %d", e.used(t, "2026-10-13"))
	}
	if _, err := e.svc.OfferToBuyer(ctx, y.Offers[1].ID); !errors.Is(err, bakery.ErrState) {
		t.Errorf("tawaran kedua saat ada yang aktif harus ditolak, dapat %v", err)
	}
	if err := e.svc.RespondOffer(ctx, y.Offers[0].ID, true); err != nil {
		t.Fatalf("RespondOffer: %v", err)
	}
	d, err := e.svc.GetOrder(ctx, y.OrderID)
	if err != nil || d.Status != "terjadwal" || d.DeliverAt != "2026-10-13 11:00" || *d.ProductionDay != "2026-10-13" {
		t.Fatalf("order setelah setuju: %+v err=%v", d, err)
	}
	if e.used(t, "2026-10-13") != 250 {
		t.Errorf("reservasi tetap harus 250, dapat %d", e.used(t, "2026-10-13"))
	}
}

func TestNegotiationExpiry(t *testing.T) {
	e := setup(t, 400, 40)
	*e.now = at(12, 9, 0)
	e.order(t, 400, at(13, 8, 0))
	y := e.order(t, 250, at(13, 8, 0))
	ctx := context.Background()
	if _, err := e.svc.OfferToBuyer(ctx, y.Offers[0].ID); err != nil {
		t.Fatal(err)
	}
	*e.now = at(12, 10, 1) // lewat batas 60 menit
	if err := e.svc.RespondOffer(ctx, y.Offers[0].ID, true); !errors.Is(err, bakery.ErrExpired) {
		t.Fatalf("harusnya bakery.ErrExpired, dapat %v", err)
	}
	if e.used(t, "2026-10-13") != 0 {
		t.Errorf("tahanan harus dilepas, terpakai %d", e.used(t, "2026-10-13"))
	}
}

func TestRejectWhenNothingFits(t *testing.T) {
	e := setup(t, 400, 40)
	*e.now = at(12, 9, 0)
	e.order(t, 400, at(13, 14, 0))
	e.order(t, 400, at(13, 14, 0)) // memenuhi Senin dan Selasa
	r := e.order(t, 500, at(13, 14, 0))
	if r.Outcome != bakery.OutcomeRejected {
		t.Fatalf("harusnya ditolak: %+v", r)
	}
	var n int
	e.db.QueryRow(`SELECT COUNT(*) FROM orders WHERE jumlah=500`).Scan(&n)
	if n != 0 {
		t.Errorf("order yang ditolak tidak boleh tersimpan")
	}
}

// 12 order bersamaan: kapasitas tidak boleh terjual lebih dari batas.
func TestConcurrentNoOversell(t *testing.T) {
	e := setup(t, 400, 40)
	var wg sync.WaitGroup
	var mu sync.Mutex
	outcomes := map[string]int{}
	var errs []error
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := e.svc.AcceptOrder(context.Background(),
				bakery.NewOrder{Items: []bakery.Item{{ProductID: e.pid, Qty: 100}}, DeliverAt: at(13, 15, 0)})
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, err)
				return
			}
			outcomes[res.Outcome]++
		}()
	}
	wg.Wait()
	for _, err := range errs {
		t.Errorf("error saat konkuren: %v", err)
	}
	rows, _ := e.db.Query(`SELECT tanggal, SUM(jumlah) FROM capacity_reservations GROUP BY tanggal`)
	defer rows.Close()
	for rows.Next() {
		var d string
		var sum int
		rows.Scan(&d, &sum)
		if sum > 440 {
			t.Errorf("OVERSELL pada %s: %d > 440", d, sum)
		}
	}
	if outcomes[bakery.OutcomeAccepted] != 8 || outcomes[bakery.OutcomeRejected] != 4 {
		t.Errorf("harusnya 8 diterima dan 4 ditolak, dapat %v", outcomes)
	}
}

// Kasus dosen: A (50 pcs) masuk lebih dulu, lalu B (800 pcs) untuk besok.
// B pesanan besar mendadak: sistem bertanya stok bahan baku dulu tanpa
// menyimpan apa pun; bila penjual menjawab tidak cukup, B ditolak dengan alasan.
// A tidak pernah tersentuh.
func TestScenarioRushBigOrder(t *testing.T) {
	e := setup(t, 800, 0)
	ctx := context.Background()
	a := e.order(t, 50, at(13, 9, 0)) // Senin 08.00, kirim Selasa 09.00
	if a.Outcome != bakery.OutcomeAccepted || a.ProductionDay != "2026-10-12" {
		t.Fatalf("A: %+v", a)
	}
	*e.now = at(12, 9, 0)
	newB := func(ready *bool) *bakery.Result {
		res, err := e.svc.AcceptOrder(ctx, bakery.NewOrder{Items: []bakery.Item{{ProductID: e.pid, Qty: 800}},
			DeliverAt: at(13, 10, 0), StockReady: ready})
		if err != nil {
			t.Fatalf("B: %v", err)
		}
		return res
	}
	if r := newB(nil); r.Outcome != bakery.OutcomeCheckStock || r.OrderID != 0 {
		t.Fatalf("harusnya tanya stok bahan baku: %+v", r)
	}
	no, yes := false, true
	if r := newB(&no); r.Outcome != bakery.OutcomeRejected || !strings.Contains(r.Reason, "luar kota") {
		t.Fatalf("bahan baku tidak cukup -> ditolak dengan alasan: %+v", r)
	}
	// bahan cukup -> masuk alur biasa; Senin sisa 750 dan A tidak bisa digeser, jadi
	// B tidak muat di Senin (ditawari jam kirim lain atau ditolak), bukan diterima
	if r := newB(&yes); r.Outcome == bakery.OutcomeAccepted {
		t.Fatalf("kapasitas tidak cukup, B tidak boleh diterima: %+v", r)
	}
	if e.used(t, "2026-10-12") != 50 {
		t.Fatalf("A harus tetap satu-satunya di Senin: %d", e.used(t, "2026-10-12"))
	}
}

// X (400 pcs) mengisi Senin, Y (300 pcs) terpaksa Selasa. Setelah X dibatalkan,
// job per jam harus memajukan Y ke Senin.
func TestRebalancePullsForward(t *testing.T) {
	e := setup(t, 400, 40)
	*e.now = at(12, 7, 0)
	ctx := context.Background()
	x := e.order(t, 400, at(13, 15, 0))
	y := e.order(t, 300, at(13, 15, 0))
	if x.ProductionDay != "2026-10-12" || y.ProductionDay != "2026-10-13" {
		t.Fatalf("awal: X=%s Y=%s", x.ProductionDay, y.ProductionDay)
	}
	if err := e.svc.Transition(ctx, x.OrderID, "cancel", "penjual"); err != nil {
		t.Fatal(err)
	}
	if err := e.svc.RunHourly(ctx); err != nil {
		t.Fatal(err)
	}
	d, err := e.svc.GetOrder(ctx, y.OrderID)
	if err != nil || d.ProductionDay == nil || *d.ProductionDay != "2026-10-12" {
		t.Fatalf("Y harusnya dimajukan ke Senin: %+v err=%v", d, err)
	}
	var n int
	e.db.QueryRow(`SELECT COUNT(*) FROM audit_log WHERE order_id=? AND aksi='dimajukan'`, y.OrderID).Scan(&n)
	if n != 1 {
		t.Errorf("audit 'dimajukan' tidak tercatat")
	}
}

func TestAtRiskAndLate(t *testing.T) {
	e := setup(t, 400, 40)
	ctx := context.Background()
	*e.now = at(12, 6, 0)
	z := e.order(t, 100, at(13, 7, 0)) // produksi Senin
	status := func() string {
		d, err := e.svc.GetOrder(ctx, z.OrderID)
		if err != nil {
			t.Fatal(err)
		}
		return d.Status
	}
	*e.now = at(12, 15, 0) // 3 jam sebelum jam produksi selesai (18.00): belum perlu peringatan
	if err := e.svc.RunHourly(ctx); err != nil || status() != "terjadwal" {
		t.Fatalf("jam 15.00 harusnya masih terjadwal: %s %v", status(), err)
	}
	*e.now = at(12, 16, 0) // 2 jam sebelum selesai, belum mulai produksi
	if err := e.svc.RunHourly(ctx); err != nil || status() != "at_risk" {
		t.Fatalf("jam 16.00 harusnya at_risk: %s %v", status(), err)
	}
	if err := e.svc.Transition(ctx, z.OrderID, "start", "penjual"); err != nil { // at_risk masih bisa dimulai
		t.Fatalf("start dari at_risk: %v", err)
	}
	*e.now = at(13, 7, 30) // lewat jam kirim, belum selesai
	if err := e.svc.RunHourly(ctx); err != nil || status() != "terlambat" {
		t.Fatalf("harusnya terlambat: %s %v", status(), err)
	}
}

func TestMonthlyReport(t *testing.T) {
	e := setup(t, 400, 40)
	ctx := context.Background()
	e.db.Exec(`DELETE FROM monthly_reports`)
	e.db.Exec(`DELETE FROM customers`)
	r, _ := e.db.Exec(`INSERT INTO customers (nama) VALUES ('Bu Sari')`)
	cid, _ := r.LastInsertId()

	mk := func(qty int, deliver, start, finish time.Time) {
		*e.now = at(12, 8, 0)
		res, err := e.svc.AcceptOrder(ctx, bakery.NewOrder{CustomerID: &cid, Items: []bakery.Item{{ProductID: e.pid, Qty: qty}}, DeliverAt: deliver})
		if err != nil || res.Outcome != bakery.OutcomeAccepted {
			t.Fatalf("order: %+v %v", res, err)
		}
		*e.now = start // produksi hanya bisa dimulai pada hari produksinya
		if err := e.svc.Transition(ctx, res.OrderID, "start", "penjual"); err != nil {
			t.Fatal(err)
		}
		*e.now = finish
		if err := e.svc.Transition(ctx, res.OrderID, "complete", "penjual"); err != nil {
			t.Fatal(err)
		}
	}
	mk(100, at(13, 15, 0), at(12, 9, 0), at(13, 14, 0)) // produksi Senin, tepat waktu
	mk(200, at(14, 15, 0), at(13, 9, 0), at(14, 16, 0)) // produksi Selasa, terlambat 1 jam

	*e.now = at(20, 9, 0)
	rep, err := e.svc.MonthlyReport(ctx, "2026-10")
	if err != nil {
		t.Fatal(err)
	}
	if rep.Summary.Orders != 2 || rep.Summary.Pcs != 300 || rep.Summary.Revenue != 300000 {
		t.Errorf("ringkasan salah: %+v", rep.Summary)
	}
	if rep.Scheduling.OnTimePct == nil || *rep.Scheduling.OnTimePct != 50 {
		t.Errorf("ketepatan waktu harusnya 50%%: %+v", rep.Scheduling.OnTimePct)
	}
	if len(rep.Daily) != 2 || len(rep.TopProducts) != 1 || rep.TopProducts[0].Share != 100 ||
		len(rep.TopCustomers) != 1 || rep.TopCustomers[0].Name != "Bu Sari" {
		t.Errorf("rincian salah: %+v", rep)
	}
	if rep.Closed {
		t.Error("belum ditutup")
	}
	if _, err := e.svc.CloseMonth(ctx, "2026-10"); !errors.Is(err, bakery.ErrState) {
		t.Errorf("bulan berjalan tidak boleh ditutup: %v", err)
	}
	*e.now = time.Date(2026, 11, 2, 9, 0, 0, 0, wib)
	if _, err := e.svc.CloseMonth(ctx, "2026-10"); err != nil {
		t.Fatalf("tutup bulan: %v", err)
	}
	if _, err := e.svc.CloseMonth(ctx, "2026-10"); !errors.Is(err, bakery.ErrState) {
		t.Errorf("tutup dua kali harus ditolak: %v", err)
	}
	// setelah ditutup, koreksi data tidak mengubah laporan
	e.db.Exec(`UPDATE orders SET total_harga = 1`)
	snap, err := e.svc.MonthlyReport(ctx, "2026-10")
	if err != nil || !snap.Closed || snap.Summary.Revenue != 300000 {
		t.Errorf("snapshot harus tetap 300000: %+v err=%v", snap.Summary, err)
	}
}

func TestStartOnlyOnProductionDay(t *testing.T) {
	e := setup(t, 400, 40)
	ctx := context.Background()
	*e.now = at(12, 7, 0)
	r := e.order(t, 100, at(14, 15, 0)) // kirim Rabu 15.00 -> produksi paling awal Selasa
	if r.ProductionDay != "2026-10-13" {
		t.Fatalf("produksi harusnya Selasa: %s", r.ProductionDay)
	}
	if err := e.svc.Transition(ctx, r.OrderID, "start", "penjual"); !errors.Is(err, bakery.ErrState) {
		t.Fatalf("mulai produksi sebelum hari produksi harus ditolak: %v", err)
	}
	d, _ := e.svc.GetOrder(ctx, r.OrderID)
	if d.CanStart || d.StatusLabel != "Terjadwal 13 Okt" {
		t.Errorf("Senin: can_start=%v label=%q", d.CanStart, d.StatusLabel)
	}
	*e.now = at(13, 7, 0) // hari produksi
	d, _ = e.svc.GetOrder(ctx, r.OrderID)
	if !d.CanStart || d.StatusLabel != "Diproduksi hari ini" || d.Status != "terjadwal" {
		t.Errorf("hari produksi: can_start=%v label=%q status=%s", d.CanStart, d.StatusLabel, d.Status)
	}
	if err := e.svc.Transition(ctx, r.OrderID, "start", "penjual"); err != nil {
		t.Fatalf("mulai di hari produksi: %v", err)
	}
	d, _ = e.svc.GetOrder(ctx, r.OrderID)
	if d.Status != "diproduksi" || d.CanStart {
		t.Errorf("setelah mulai: status=%s can_start=%v", d.Status, d.CanStart)
	}
}

func TestBaruBecomesTerjadwalInWindow(t *testing.T) {
	e := setup(t, 400, 40)
	ctx := context.Background()
	*e.now = at(12, 8, 0)
	far := e.order(t, 100, at(18, 15, 0)) // Minggu 18 Okt: di luar jendela 72 jam
	d, _ := e.svc.GetOrder(ctx, far.OrderID)
	if d.Status != "baru" || d.StatusLabel != "Baru" || d.ProductionDay == nil {
		t.Fatalf("order jauh hari harus Baru dan tetap memegang reservasi: %+v", d)
	}
	*e.now = at(13, 9, 0)
	if err := e.svc.RunHourly(ctx); err != nil {
		t.Fatal(err)
	}
	if d, _ = e.svc.GetOrder(ctx, far.OrderID); d.Status != "baru" {
		t.Errorf("belum masuk jendela, harusnya masih Baru: %s", d.Status)
	}
	*e.now = at(15, 20, 0) // 67 jam sebelum kirim
	if err := e.svc.RunHourly(ctx); err != nil {
		t.Fatal(err)
	}
	if d, _ = e.svc.GetOrder(ctx, far.OrderID); d.Status != "terjadwal" || d.StatusLabel != "Terjadwal 17 Okt" {
		t.Errorf("harusnya Terjadwal 17 Okt: %s %q", d.Status, d.StatusLabel)
	}
}

func TestUpdateSettings(t *testing.T) {
	e := setup(t, 400, 40)
	ctx := context.Background()
	*e.now = at(12, 8, 0)
	for name, in := range map[string]map[string]float64{
		"tak dikenal":   {"hack": 1},
		"di luar batas": {"kapasitas_harian": 0},
		"bukan bulat":   {"kapasitas_harian": 800.5},
		"campur salah":  {"kapasitas_harian": 800, "jendela_jam": -3},
		"kosong":        {},
	} {
		if err := e.svc.UpdateSettings(ctx, in); !errors.Is(err, bakery.ErrValidation) {
			t.Errorf("%s: harusnya ErrValidation, dapat %v", name, err)
		}
	}
	m, _ := e.svc.SettingsMap(ctx)
	if m["kapasitas_harian"] != 400 {
		t.Fatalf("validasi gagal tidak boleh menyimpan sebagian: %v", m["kapasitas_harian"])
	}
	e.order(t, 300, at(13, 15, 0)) // reservasi 300 di Senin
	if err := e.svc.UpdateSettings(ctx, map[string]float64{"kapasitas_harian": 200, "lembur_maks": 0}); !errors.Is(err, bakery.ErrState) {
		t.Fatalf("kapasitas di bawah reservasi harus ditolak: %v", err)
	}
	if err := e.svc.UpdateSettings(ctx, map[string]float64{"kapasitas_harian": 800, "lembur_maks": 80}); err != nil {
		t.Fatal(err)
	}
	var base int
	e.db.QueryRow(`SELECT kapasitas_dasar FROM daily_capacity WHERE tanggal='2026-10-12'`).Scan(&base)
	if base != 800 {
		t.Errorf("kapasitas hari mendatang harus ikut berubah: %d", base)
	}
	var n int
	e.db.QueryRow(`SELECT COUNT(*) FROM audit_log WHERE aksi='ubah_pengaturan'`).Scan(&n)
	if n != 1 {
		t.Errorf("audit pengaturan: %d", n)
	}
}

func TestUpdateProductAndExport(t *testing.T) {
	e := setup(t, 400, 40)
	ctx := context.Background()
	nama, harga, aktif := "Roti Manis", 2000.0, false
	if err := e.svc.UpdateProduct(ctx, e.pid, &nama, &harga, &aktif); err != nil {
		t.Fatal(err)
	}
	if err := e.svc.UpdateProduct(ctx, 999999, &nama, nil, nil); !errors.Is(err, bakery.ErrNotFound) {
		t.Errorf("produk tak ada: %v", err)
	}
	if err := e.svc.UpdateProduct(ctx, e.pid, nil, nil, nil); !errors.Is(err, bakery.ErrValidation) {
		t.Errorf("tanpa perubahan: %v", err)
	}
	// produk nonaktif tidak bisa dipesan
	*e.now = at(12, 8, 0)
	if _, err := e.svc.AcceptOrder(ctx, bakery.NewOrder{Items: []bakery.Item{{ProductID: e.pid, Qty: 10}}, DeliverAt: at(13, 15, 0)}); !errors.Is(err, bakery.ErrValidation) {
		t.Errorf("produk nonaktif harus ditolak: %v", err)
	}
	b, _, err := e.svc.ExportMonthlyCSV(ctx, "2026-10")
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if len(b) < 3 || b[0] != 0xEF || !strings.Contains(s, "RINGKASAN") || !strings.Contains(s, "Laporan Penjualan Bulanan;2026-10") {
		t.Errorf("CSV tidak sesuai: %q", s[:min(len(s), 120)])
	}
}

func TestAuditLogIncludesData(t *testing.T) {
	e := setup(t, 400, 40)
	ctx := context.Background()
	*e.now = at(12, 8, 0)
	r := e.order(t, 100, at(13, 15, 0))
	logs, err := e.svc.AuditLog(ctx, 20)
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range logs {
		if l.Aksi == "reservasi" && l.OrderID != nil && *l.OrderID == r.OrderID && l.Sesudah != nil {
			return
		}
	}
	t.Errorf("entri reservasi dengan order_id dan data sesudah tidak ditemukan: %v", logs)
}
