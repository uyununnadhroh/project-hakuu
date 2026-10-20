package bakery

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"bakery-scheduler/internal/entity/bakery"

	"github.com/jmoiron/sqlx"
)

type line struct {
	ProductID int64
	Qty       int
	Price     float64
	Variants  []bakery.ItemVariant
}

type orderRef struct {
	ID        int64
	Kode      string
	Qty       int
	DeliverAt time.Time
	Actor     string
	InWindow  bool
}

type moveInternal struct {
	ResID, OrderID int64
	Kode           string
	From, To       time.Time
	Qty            int
	OldScore       float64
}

type plan struct {
	Day      time.Time
	Moves    []moveInternal
	Overtime int
}

// parseTime menerima "YYYY-MM-DD HH:MM", "YYYY-MM-DDTHH:MM", atau RFC3339.
func (s Service) parseTime(v string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04", "2006-01-02 15:04"} {
		if t, err := time.ParseInLocation(layout, v, s.loc); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("%w: format waktu harus YYYY-MM-DD HH:MM", bakery.ErrValidation)
}

// CreateOrder dipanggil dari POST /api/orders: siapkan pelanggan lalu jalankan AcceptOrder.
// actor = user yang sedang login; dicatat sebagai penanggung jawab order.
func (s Service) CreateOrder(ctx context.Context, req bakery.ReqCreateOrder, actor bakery.Actor) (*bakery.Result, error) {
	t, err := s.parseTime(req.DeliverAt)
	if err != nil {
		return nil, err
	}
	if req.CustomerID == nil && req.Customer != nil {
		cid, err := s.FindOrCreateCustomer(ctx, strings.TrimSpace(req.Customer.Nama),
			strings.TrimSpace(req.Customer.Telepon), strings.TrimSpace(req.Customer.Alamat))
		if err != nil {
			return nil, err
		}
		req.CustomerID = &cid
	}
	return s.AcceptOrder(ctx, bakery.NewOrder{CustomerID: req.CustomerID, Items: req.Items, DeliverAt: t, Note: req.Note,
		Actor: actor.Nama, CreatedBy: actor.UserID, StockReady: req.BahanSiap})
}

// priceItems menghitung harga dan beban kapasitas. Qty tiap item adalah jumlah
// BOX; beban kapasitas = jumlah box x bobot produk (bertopping lebih berat),
// dibulatkan ke atas. Kembalian: (jumlah box, beban kapasitas, total harga, baris, error).
func (s Service) priceItems(ctx context.Context, tx *sqlx.Tx, items []bakery.Item) (int, int, float64, []line, error) {
	var boxes int
	var load, total float64
	var lines []line
	for _, it := range items {
		if it.Qty <= 0 {
			return 0, 0, 0, nil, fmt.Errorf("%w: jumlah item harus lebih dari 0", bakery.ErrValidation)
		}
		p, err := s.bakery.GetProductPrice(ctx, tx, it.ProductID)
		if errors.Is(err, sql.ErrNoRows) || (err == nil && !p.Aktif) {
			return 0, 0, 0, nil, fmt.Errorf("%w: produk %d tidak tersedia", bakery.ErrValidation, it.ProductID)
		}
		if err != nil {
			return 0, 0, 0, nil, err
		}
		var avail []bakery.ProductVariant
		if p.MaksVarian > 0 {
			if avail, err = s.bakery.GetVariantsByProduct(ctx, tx, it.ProductID); err != nil {
				return 0, 0, 0, nil, err
			}
		}
		if err := CheckVariants(p.MaksVarian, avail, it.Variants); err != nil {
			return 0, 0, 0, nil, fmt.Errorf("%w: produk %d: %v", bakery.ErrValidation, it.ProductID, err)
		}
		boxes += it.Qty
		load += float64(it.Qty) * p.Bobot
		total += p.Harga * float64(it.Qty)
		lines = append(lines, line{it.ProductID, it.Qty, p.Harga, it.Variants})
	}
	return boxes, int(math.Ceil(load - 1e-9)), total, lines, nil
}

// AcceptOrder menjalankan alur penerimaan order dalam SATU transaksi:
// cek pesanan besar mendadak -> reservasi langsung -> geser order lain ->
// lembur -> rekomendasi jam kirim -> tolak.
//
// Pesanan besar mendadak (lihat IsRushBigOrder) tidak langsung dijadwalkan:
// bila penjual belum menjawab (in.StockReady nil) hasilnya perlu_cek_bahan dan
// tidak ada yang disimpan; frontend mengirim ulang order dengan jawaban penjual.
func (s Service) AcceptOrder(ctx context.Context, in bakery.NewOrder) (*bakery.Result, error) {
	return retryDeadlock(func() (*bakery.Result, error) { return s.acceptOnce(ctx, in) })
}

func (s Service) acceptOnce(ctx context.Context, in bakery.NewOrder) (*bakery.Result, error) {
	now := s.now()
	in.DeliverAt = in.DeliverAt.In(s.loc)
	if in.Actor == "" {
		in.Actor = "penjual"
	}
	if len(in.Items) == 0 {
		return nil, fmt.Errorf("%w: item order kosong", bakery.ErrValidation)
	}
	if !in.DeliverAt.After(now) {
		return nil, fmt.Errorf("%w: jam kirim harus di masa depan", bakery.ErrValidation)
	}

	pre, err := s.loadSettings(ctx, nil)
	if err != nil {
		return nil, err
	}
	if err := CheckDeliveryHour(in.DeliverAt, pre); err != nil {
		return nil, err
	}
	deliverDay0 := DayOf(in.DeliverAt)
	if err := s.ensureDays(ctx, AddDays(deliverDay0, -1), deliverDay0, pre); err != nil {
		return nil, err
	}

	tx, err := s.bakery.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	set, err := s.loadSettings(ctx, tx)
	if err != nil {
		return nil, err
	}
	boxes, qty, total, lines, err := s.priceItems(ctx, tx, in.Items)
	if err != nil {
		return nil, err
	}

	// 0. Pesanan besar mendadak: tanyakan stok bahan baku ke penjual dulu.
	rush := IsRushBigOrder(qty, in.DeliverAt, now, set)
	if rush && in.StockReady == nil {
		_ = tx.Rollback()
		return &bakery.Result{Outcome: bakery.OutcomeCheckStock, Reason: fmt.Sprintf(
			"Pesanan %d pcs termasuk pesanan besar (di atas %d pcs) dan dipesan kurang dari %d hari sebelum pengiriman. "+
				"Bahan baku dibeli di luar kota, pastikan stok bahan baku cukup sebelum menerima pesanan ini.",
			qty, set.BigOrderQty, set.BigOrderLeadDays)}, nil
	}
	if rush && !*in.StockReady {
		return s.reject(ctx, tx, in, qty, fmt.Sprintf(
			"Pesanan di atas %d pcs harus dipesan paling lambat %d hari sebelum pengiriman, karena bahan baku "+
				"dibeli di luar kota dan tidak tersedia di toko sekitar. Stok bahan baku saat ini tidak mencukupi.",
			set.BigOrderQty, set.BigOrderLeadDays), nil)
	}

	// Hari produksi yang boleh: H-1, lalu hari H (hanya bila jam kirim >= jam siap 11.00).
	deadline := DeadlineDay(in.DeliverAt, set)
	earliest := EarliestDay(in.DeliverAt, now, set)
	// H-1 sudah lewat dan jam kirim sebelum jam siap: tidak ada hari langsung,
	// tapi masih bisa dinegosiasikan ke jam >= 11.00 dengan produksi hari H (langkah 4).
	noDirectDay := earliest.After(deadline)

	// Kunci semua hari yang mungkin terlibat, selalu urut naik (hindari deadlock).
	deliverDay, today := DayOf(in.DeliverAt), DayOf(now)
	for d := AddDays(deliverDay, -1); !d.After(deliverDay); d = AddDays(d, 1) {
		if d.Before(today) {
			continue
		}
		if _, err := s.getDay(ctx, tx, d, set, now); err != nil {
			return nil, err
		}
	}

	id, err := s.bakery.InsertOrder(ctx, tx, bakery.OrderInsert{
		CustomerID: in.CustomerID, Jumlah: qty, QtyBox: boxes, TotalHarga: total,
		WaktuPesan: now, WaktuKirim: in.DeliverAt, HariDeadline: fmtDay(deadline), Catatan: in.Note,
		CreatedBy: in.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	o := orderRef{ID: id, Kode: fmt.Sprintf("ORD-%05d", id), Qty: qty, DeliverAt: in.DeliverAt, Actor: in.Actor,
		InWindow: InWindow(in.DeliverAt, now, set)}
	if err := s.bakery.UpdateOrderKode(ctx, tx, id, o.Kode); err != nil {
		return nil, err
	}
	if rush {
		if err := s.audit(ctx, tx, id, o.Actor, "pesanan_besar_mendadak", nil, map[string]int{"jumlah": qty},
			"Pesanan besar dipesan mendadak; penjual menyatakan stok bahan baku cukup"); err != nil {
			return nil, err
		}
	}
	for _, l := range lines {
		itemID, err := s.bakery.InsertOrderItem(ctx, tx, id, l.ProductID, l.Qty, l.Price)
		if err != nil {
			return nil, err
		}
		for _, v := range l.Variants {
			if err := s.bakery.InsertOrderItemVariant(ctx, tx, itemID, v.VariantID, v.Opsi); err != nil {
				return nil, err
			}
		}
	}
	sc := ComputeScore(bakery.ScoreInput{Qty: qty, DeliverAt: in.DeliverAt, OrderedAt: now}, now, set)

	if !noDirectDay {
		// 1. Cek H-1 dulu; bila penuh dan jam kirim >= 11.00, langsung dialihkan ke hari H.
		for d := earliest; !d.After(deadline); d = AddDays(d, 1) {
			st, err := s.getDay(ctx, tx, d, set, now)
			if err != nil {
				return nil, err
			}
			if st.Free() >= qty {
				return s.finishAccept(ctx, tx, o, d, 0, nil, sc, now)
			}
		}

		// 2-3. Geser order berskor lebih rendah (tetap di H-1/hari H miliknya), lalu lembur.
		p, err := s.resolveConflict(ctx, tx, sc.Total, qty, earliest, deadline, set, now)
		if err != nil {
			return nil, err
		}
		if p != nil {
			return s.finishAccept(ctx, tx, o, p.Day, p.Overtime, p.Moves, sc, now)
		}
	}

	// 4. H-1 tidak bisa: tawarkan jam kirim lain di hari yang sama (mulai jam siap 11.00,
	//    produksi dialihkan ke hari H).
	var ferr error
	free := func(day time.Time) int {
		if ferr != nil {
			return 0
		}
		st, err := s.getDay(ctx, tx, day, set, now)
		if err != nil {
			ferr = err
			return 0
		}
		return st.Free()
	}
	alts := AlternativeTimes(in.DeliverAt, qty, now, set, free)
	if ferr != nil {
		return nil, ferr
	}
	if len(alts) > 0 {
		out := &bakery.Result{Outcome: bakery.OutcomeNegotiation, OrderID: id, Kode: o.Kode,
			Reason: fmt.Sprintf("Kapasitas produksi H-1 tidak cukup; produksi bisa dialihkan ke hari kirim bila jam kirim mulai %s", ReadyLabel(set))}
		for i, a := range alts {
			oid, err := s.bakery.InsertOffer(ctx, tx, id, i+1, a.DeliverAt, fmtDay(a.ProductionDay))
			if err != nil {
				return nil, err
			}
			out.Offers = append(out.Offers, bakery.Offer{ID: oid, Seq: i + 1,
				DeliverAt: a.DeliverAt.Format("2006-01-02 15:04"), ProductionDay: fmtDay(a.ProductionDay)})
		}
		if err := s.bakery.UpdateOrderNegotiation(ctx, tx, id, sc, now); err != nil {
			return nil, err
		}
		if err := s.audit(ctx, tx, id, o.Actor, "perlu_negosiasi", nil, out.Offers, out.Reason); err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return out, nil
	}

	// 5. Tidak ada jam yang muat di tanggal itu (sampai 20.00): tolak.
	//    Rekomendasi hanya di hari yang sama, jadi tidak ada saran tanggal lain.
	if noDirectDay {
		return s.reject(ctx, tx, in, qty, "Tidak sempat diproduksi sebelum jam kirim, dan tidak ada jam kirim lain yang muat di hari itu", nil)
	}
	return s.reject(ctx, tx, in, qty, fmt.Sprintf("Kapasitas tidak tersedia untuk jam kirim %s–%s di hari itu",
		ReadyLabel(set), fmt.Sprintf("%02d.00", set.LastDeliveryHour)), nil)
}

// reject membatalkan seluruh transaksi (order tidak tersimpan) lalu mencatat
// penolakan di audit log.
func (s Service) reject(ctx context.Context, tx *sqlx.Tx, in bakery.NewOrder, qty int, reason string, dates []string) (*bakery.Result, error) {
	_ = tx.Rollback()
	_ = s.audit(ctx, nil, 0, in.Actor, "order_ditolak", nil,
		map[string]interface{}{"jumlah": qty, "waktu_kirim": in.DeliverAt.Format(time.RFC3339)}, reason)
	return &bakery.Result{Outcome: bakery.OutcomeRejected, Reason: reason, AltDates: dates}, nil
}

func (s Service) finishAccept(ctx context.Context, tx *sqlx.Tx, o orderRef, day time.Time, overtime int,
	moves []moveInternal, sc bakery.Score, now time.Time) (*bakery.Result, error) {

	out := &bakery.Result{Outcome: bakery.OutcomeAccepted, OrderID: o.ID, Kode: o.Kode, ProductionDay: fmtDay(day), Overtime: overtime}
	for _, m := range moves {
		if err := s.bakery.UpdateReservationDay(ctx, tx, m.ResID, fmtDay(m.To)); err != nil {
			return nil, err
		}
		if err := s.bakery.UpdateOrderProductionDay(ctx, tx, m.OrderID, fmtDay(m.To)); err != nil {
			return nil, err
		}
		reason := fmt.Sprintf("Digeser karena %s (skor %.3f) lebih prioritas dari skor %.3f; deadline tetap terpenuhi",
			o.Kode, sc.Total, m.OldScore)
		if err := s.audit(ctx, tx, m.OrderID, "sistem", "digeser",
			map[string]string{"hari_produksi": fmtDay(m.From)}, map[string]string{"hari_produksi": fmtDay(m.To)}, reason); err != nil {
			return nil, err
		}
		if err := s.notify(ctx, tx, m.OrderID, fmt.Sprintf("Order %s digeser dari %s ke %s karena ada order prioritas (%s).",
			m.Kode, fmtDay(m.From), fmtDay(m.To), o.Kode)); err != nil {
			return nil, err
		}
		out.Moved = append(out.Moved, bakery.Move{OrderID: m.OrderID, Kode: m.Kode, From: fmtDay(m.From), To: fmtDay(m.To), Qty: m.Qty})
	}
	if _, err := s.bakery.InsertReservation(ctx, tx, o.ID, fmtDay(day), o.Qty, overtime, "tetap", nil); err != nil {
		return nil, err
	}
	// Di luar jendela 72 jam: hanya memegang reservasi (status Baru).
	// Otomatis menjadi Terjadwal oleh job per jam saat masuk jendela.
	status := "terjadwal"
	if !o.InWindow {
		status = "baru"
	}
	if err := s.bakery.UpdateOrderAccepted(ctx, tx, o.ID, status, fmtDay(day), overtime, sc, now); err != nil {
		return nil, err
	}
	if err := s.audit(ctx, tx, o.ID, o.Actor, "reservasi", nil,
		map[string]interface{}{"hari_produksi": fmtDay(day), "jumlah": o.Qty, "lembur": overtime}, "Reservasi kapasitas"); err != nil {
		return nil, err
	}
	if overtime > 0 {
		if err := s.audit(ctx, tx, o.ID, "sistem", "lembur", nil, map[string]int{"pcs": overtime},
			"Kapasitas normal penuh, memakai lembur"); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}

// resolveConflict mensimulasikan penggeseran di memori. Tidak ada yang ditulis
// ke database; finishAccept menerapkannya hanya bila ada rencana yang berhasil.
func (s Service) resolveConflict(ctx context.Context, tx *sqlx.Tx, newScore float64, qty int,
	earliest, deadline time.Time, set bakery.Settings, now time.Time) (*plan, error) {

	for day := earliest; !day.After(deadline); day = AddDays(day, 1) {
		st, err := s.getDay(ctx, tx, day, set, now)
		if err != nil {
			return nil, err
		}
		var cands []bakery.Candidate
		if !st.Locked { // hari yang dikunci (frozen window) tidak boleh diacak
			if cands, err = s.loadCandidates(ctx, tx, day, newScore, set, now); err != nil {
				return nil, err
			}
		}
		delta := map[string]int{}
		free := st.Base - st.Used // bisa negatif bila sudah ada lembur
		moved := 0
		var moves []moveInternal
		for _, c := range cands {
			if free >= qty {
				break
			}
			to, ok, err := s.findReplacement(ctx, tx, c, day, delta, set, now)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue // menggeser order ini akan membuatnya terlambat
			}
			moves = append(moves, moveInternal{c.ResID, c.OrderID, c.Kode, day, to, c.Qty, c.Score})
			delta[fmtDay(to)] += c.Qty
			free += c.Qty
			moved += c.Qty
		}
		if free >= qty {
			return &plan{Day: day, Moves: moves}, nil
		}
		usedAfter := st.Used - moved
		if qty <= st.Base+st.OvertimeMax-usedAfter {
			return &plan{Day: day, Moves: moves, Overtime: qty - max(0, st.Base-usedAfter)}, nil
		}
	}
	return nil, nil
}

func (s Service) loadCandidates(ctx context.Context, tx *sqlx.Tx, day time.Time, newScore float64,
	set bakery.Settings, now time.Time) ([]bakery.Candidate, error) {

	all, err := s.bakery.GetCandidates(ctx, tx, fmtDay(day))
	if err != nil {
		return nil, err
	}
	var out []bakery.Candidate
	for _, c := range all {
		if c.Deadline, err = s.parseDay(c.DeadlineStr); err != nil {
			return nil, err
		}
		c.Score = ComputeScore(bakery.ScoreInput{Qty: c.Qty, DeliverAt: c.DeliverAt, OrderedAt: c.Ordered}, now, set).Total
		if c.Score <= newScore-set.MinScoreGap { // ambang anti-reshuffle
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Score < out[j].Score })
	return out, nil
}

func (s Service) findReplacement(ctx context.Context, tx *sqlx.Tx, c bakery.Candidate, from time.Time,
	delta map[string]int, set bakery.Settings, now time.Time) (time.Time, bool, error) {

	for d := EarliestDay(c.DeliverAt, now, set); !d.After(c.Deadline); d = AddDays(d, 1) {
		if d.Equal(from) {
			continue
		}
		st, err := s.getDay(ctx, tx, d, set, now)
		if err != nil {
			return time.Time{}, false, err
		}
		if st.Base-(st.Used+delta[fmtDay(d)]) >= c.Qty {
			return d, true, nil
		}
	}
	return time.Time{}, false, nil
}
