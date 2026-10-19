package bakery

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"bakery-scheduler/internal/entity/bakery"
)

// parseMonth "2026-10" -> awal bulan. Kosong = bulan berjalan.
func (s Service) parseMonth(v string) (time.Time, error) {
	if v == "" {
		v = s.now().Format("2006-01")
	}
	t, err := time.ParseInLocation("2006-01", v, s.loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: format bulan harus YYYY-MM", bakery.ErrValidation)
	}
	return t, nil
}

func (s Service) summary(ctx context.Context, from, to time.Time) (bakery.Summary, error) {
	m, err := s.bakery.GetSummary(ctx, from, to)
	if m.Orders > 0 {
		m.AvgOrder = m.Revenue / float64(m.Orders)
	}
	return m, err
}

func (s Service) computeReport(ctx context.Context, month time.Time) (*bakery.Report, error) {
	from := month
	to := from.AddDate(0, 1, 0)
	r := &bakery.Report{Month: from.Format("2006-01")}
	var err error
	if r.Summary, err = s.summary(ctx, from, to); err != nil {
		return nil, err
	}
	if r.Previous, err = s.summary(ctx, from.AddDate(0, -1, 0), from); err != nil {
		return nil, err
	}
	if r.Previous.Revenue > 0 {
		v := (r.Summary.Revenue - r.Previous.Revenue) / r.Previous.Revenue * 100
		r.RevenueChangePct = &v
	}

	// tren harian (berdasarkan tanggal kirim)
	if r.Daily, err = s.bakery.GetDailyRevenue(ctx, from, to); err != nil {
		return nil, err
	}
	// produk terlaris
	if r.TopProducts, err = s.bakery.GetTopProducts(ctx, from, to); err != nil {
		return nil, err
	}
	for i := range r.TopProducts {
		if r.Summary.Revenue > 0 {
			r.TopProducts[i].Share = r.TopProducts[i].Revenue / r.Summary.Revenue * 100
		}
	}
	// pelanggan teratas
	if r.TopCustomers, err = s.bakery.GetTopCustomers(ctx, from, to); err != nil {
		return nil, err
	}
	// status order (batal tampil di sini, tidak masuk penjualan)
	if r.Status, err = s.bakery.GetStatusCounts(ctx, from, to); err != nil {
		return nil, err
	}

	// kinerja penjadwalan
	sc := &r.Scheduling
	if sc.Delivered, sc.OnTime, err = s.bakery.GetDeliveredStats(ctx, from, to); err != nil {
		return nil, err
	}
	if sc.Delivered > 0 {
		v := float64(sc.OnTime) / float64(sc.Delivered) * 100
		sc.OnTimePct = &v
	}
	util, err := s.bakery.GetUtilization(ctx, fmtDay(from), fmtDay(to))
	if err != nil {
		return nil, err
	}
	if util.Valid {
		v := util.Float64
		sc.UtilizationPct = &v
	}
	if sc.OvertimeDays, sc.OvertimePcs, err = s.bakery.GetOvertimeStats(ctx, fmtDay(from), fmtDay(to)); err != nil {
		return nil, err
	}
	counts, err := s.bakery.GetAuditCounts(ctx, from, to)
	if err != nil {
		return nil, err
	}
	for _, c := range counts {
		switch c.Aksi {
		case "digeser":
			sc.Moved = c.Total
		case "dimajukan":
			sc.Advanced = c.Total
		case "negosiasi_ditawarkan":
			sc.NegOffered = c.Total
		case "negosiasi_diterima":
			sc.NegAccepted = c.Total
		case "negosiasi_ditolak", "negosiasi_kedaluwarsa":
			sc.NegDeclined += c.Total
		case "order_ditolak":
			sc.Rejected = c.Total
		}
	}
	return r, nil
}

// MonthlyReport: bulan yang sudah ditutup dibaca dari snapshot; selain itu dihitung langsung.
func (s Service) MonthlyReport(ctx context.Context, monthStr string) (*bakery.Report, error) {
	month, err := s.parseMonth(monthStr)
	if err != nil {
		return nil, err
	}
	raw, at, err := s.bakery.GetMonthlyReport(ctx, month.Format("2006-01"))
	if err == nil {
		var r bakery.Report
		if err := json.Unmarshal([]byte(raw), &r); err != nil {
			return nil, err
		}
		r.Closed, r.ClosedAt = true, at.Format("2006-01-02 15:04")
		return &r, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return s.computeReport(ctx, month)
}

// CloseMonth menyimpan snapshot laporan. Hanya bulan yang sudah berakhir.
func (s Service) CloseMonth(ctx context.Context, monthStr string) (*bakery.Report, error) {
	month, err := s.parseMonth(monthStr)
	if err != nil {
		return nil, err
	}
	if s.now().Before(month.AddDate(0, 1, 0)) {
		return nil, fmt.Errorf("%w: bulan ini belum berakhir", bakery.ErrState)
	}
	exists, err := s.bakery.CountMonthlyReport(ctx, month.Format("2006-01"))
	if err != nil {
		return nil, err
	}
	if exists > 0 {
		return nil, fmt.Errorf("%w: bulan ini sudah ditutup", bakery.ErrState)
	}
	r, err := s.computeReport(ctx, month)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	if err := s.bakery.InsertMonthlyReport(ctx, r.Month, string(raw)); err != nil {
		return nil, err
	}
	_ = s.audit(ctx, nil, 0, "penjual", "tutup_bulan", nil, map[string]string{"bulan": r.Month}, "Snapshot laporan bulanan disimpan")
	r.Closed = true
	return r, nil
}

// ExportMonthlyCSV menghasilkan CSV berformat Indonesia agar langsung terbuka
// rapi di Excel: pemisah titik koma, desimal koma, rupiah tanpa pemisah ribuan,
// dan BOM UTF-8 supaya huruf tidak rusak. Mengembalikan (isi file, bulan).
func (s Service) ExportMonthlyCSV(ctx context.Context, month string) ([]byte, string, error) {
	r, err := s.MonthlyReport(ctx, month)
	if err != nil {
		return nil, "", err
	}
	var buf bytes.Buffer
	buf.WriteString("\xEF\xBB\xBF")
	w := csv.NewWriter(&buf)
	w.Comma = ';'
	rp := func(v float64) string { return strconv.FormatFloat(v, 'f', 0, 64) }
	dec := func(v float64) string { return strings.Replace(strconv.FormatFloat(v, 'f', 1, 64), ".", ",", 1) }
	pct := func(v *float64) string {
		if v == nil {
			return "-"
		}
		return dec(*v)
	}
	row := func(c ...string) { _ = w.Write(c) }
	blank := func() { row() }

	status := "belum ditutup"
	if r.Closed {
		status = "ditutup " + r.ClosedAt
	}
	row("Laporan Penjualan Bulanan", r.Month, status)
	blank()
	row("RINGKASAN")
	row("Total pendapatan (Rp)", rp(r.Summary.Revenue))
	row("Jumlah order", strconv.Itoa(r.Summary.Orders))
	row("Total pcs terjual", strconv.Itoa(r.Summary.Pcs))
	row("Rata-rata nilai order (Rp)", rp(r.Summary.AvgOrder))
	row("Pendapatan bulan lalu (Rp)", rp(r.Previous.Revenue))
	row("Perubahan pendapatan (%)", pct(r.RevenueChangePct))
	blank()
	row("TREN HARIAN")
	row("Tanggal", "Order", "Pcs", "Pendapatan (Rp)")
	for _, d := range r.Daily {
		row(d.Date, strconv.Itoa(d.Orders), strconv.Itoa(d.Pcs), rp(d.Revenue))
	}
	blank()
	row("PRODUK TERLARIS")
	row("Produk", "Pcs", "Pendapatan (Rp)", "Kontribusi (%)")
	for _, p := range r.TopProducts {
		row(p.Name, strconv.Itoa(p.Pcs), rp(p.Revenue), dec(p.Share))
	}
	blank()
	row("PELANGGAN TERATAS")
	row("Pelanggan", "Order", "Total belanja (Rp)")
	for _, c := range r.TopCustomers {
		row(c.Name, strconv.Itoa(c.Orders), rp(c.Spend))
	}
	blank()
	row("STATUS ORDER")
	for _, st := range []string{"selesai", "batal", "terlambat", "diproduksi", "terjadwal", "baru", "at_risk", "menunggu_negosiasi"} {
		if n, ok := r.Status[st]; ok {
			row(st, strconv.Itoa(n))
		}
	}
	blank()
	row("KINERJA PENJADWALAN")
	sc := r.Scheduling
	row("Order selesai diserahkan", strconv.Itoa(sc.Delivered))
	row("Tepat waktu", strconv.Itoa(sc.OnTime))
	row("Ketepatan waktu (%)", pct(sc.OnTimePct))
	row("Utilisasi kapasitas (%)", pct(sc.UtilizationPct))
	row("Hari lembur", strconv.Itoa(sc.OvertimeDays))
	row("Pcs lembur", strconv.Itoa(sc.OvertimePcs))
	row("Order digeser", strconv.Itoa(sc.Moved))
	row("Order dimajukan", strconv.Itoa(sc.Advanced))
	row("Negosiasi jam ditawarkan", strconv.Itoa(sc.NegOffered))
	row("Negosiasi diterima", strconv.Itoa(sc.NegAccepted))
	row("Negosiasi ditolak/kedaluwarsa", strconv.Itoa(sc.NegDeclined))
	row("Order ditolak", strconv.Itoa(sc.Rejected))
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, "", fmt.Errorf("gagal menulis CSV: %w", err)
	}
	return buf.Bytes(), r.Month, nil
}
