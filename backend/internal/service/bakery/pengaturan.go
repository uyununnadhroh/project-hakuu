package bakery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"

	"bakery-scheduler/internal/entity/bakery"
)

type settingRule struct {
	Min, Max float64
	Int      bool
}

// Daftar putih parameter yang boleh diubah lewat API beserta batas wajarnya.
var settingRules = map[string]settingRule{
	"kapasitas_harian":               {1, 100000, true},
	"lembur_maks":                    {0, 100000, true},
	"jendela_jam":                    {1, 720, false},
	"batas_pagi_jam":                 {0, 23, true},
	"lama_menunggu_maks_jam":         {1, 720, false},
	"pembulatan_jam":                 {1, 12, true},
	"ambang_selisih_skor":            {0, 1, false},
	"maks_tawaran_jam":               {1, 10, true},
	"batas_respons_menit":            {1, 1440, true},
	"jam_produksi_selesai":           {0, 23, true},
	"menit_packing":                  {0, 600, true},
	"menit_tempuh":                   {0, 600, true},
	"jam_kirim_terakhir":             {0, 23, true},
	"jam_peringatan_sebelum_selesai": {0, 12, true},
}

// SettingsMap semua parameter mentah (kunci -> nilai).
func (s Service) SettingsMap(ctx context.Context) (map[string]float64, error) {
	return s.bakery.GetSettings(ctx, nil)
}

func (s Service) Settings(ctx context.Context) (bakery.Settings, error) {
	return s.loadSettings(ctx, nil)
}

// GetSettingsView isi GET /api/settings: {parameters, weights}.
func (s Service) GetSettingsView(ctx context.Context) (map[string]interface{}, error) {
	params, err := s.bakery.GetSettings(ctx, nil)
	if err != nil {
		return nil, err
	}
	set, err := s.loadSettings(ctx, nil)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"parameters": params,
		"weights":    map[string]float64{"J": set.Weights.J, "Q": set.Weights.Q, "W": set.Weights.W},
	}, nil
}

// UpdateSettings memvalidasi SEMUA nilai dulu; bila ada satu yang salah, tidak
// ada yang disimpan. Perubahan kapasitas langsung berlaku untuk hari-hari
// mendatang yang barisnya sudah dibuat, dan ditolak bila lebih kecil dari
// reservasi yang sudah ada.
func (s Service) UpdateSettings(ctx context.Context, in map[string]float64) error {
	if len(in) == 0 {
		return fmt.Errorf("%w: tidak ada parameter yang diubah", bakery.ErrValidation)
	}
	keys := make([]string, 0, len(in))
	for k, v := range in {
		r, ok := settingRules[k]
		if !ok {
			return fmt.Errorf("%w: parameter %q tidak dikenal", bakery.ErrValidation, k)
		}
		if math.IsNaN(v) || v < r.Min || v > r.Max {
			return fmt.Errorf("%w: %s harus antara %g dan %g", bakery.ErrValidation, k, r.Min, r.Max)
		}
		if r.Int && v != math.Trunc(v) {
			return fmt.Errorf("%w: %s harus bilangan bulat", bakery.ErrValidation, k)
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	now := s.now()
	tx, err := s.bakery.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	cur, err := s.bakery.GetSettings(ctx, tx)
	if err != nil {
		return err
	}
	merged := map[string]float64{}
	for k, v := range cur {
		merged[k] = v
	}
	for k, v := range in {
		merged[k] = v
	}

	_, capChanged := in["kapasitas_harian"]
	_, otChanged := in["lembur_maks"]
	if capChanged || otChanged {
		base, ot := merged["kapasitas_harian"], merged["lembur_maks"]
		today := fmtDay(now)
		over, err := s.bakery.CountOverReserved(ctx, tx, today, now, base+ot)
		if err != nil {
			return err
		}
		if over > 0 {
			return fmt.Errorf("%w: kapasitas baru lebih kecil dari reservasi yang sudah ada di %d hari", bakery.ErrState, over)
		}
		if err := s.bakery.UpdateFutureCapacity(ctx, tx, int(base), int(ot), today); err != nil {
			return err
		}
	}

	before, after := map[string]float64{}, map[string]float64{}
	for _, k := range keys {
		before[k], after[k] = cur[k], in[k]
		if err := s.bakery.UpsertSetting(ctx, tx, k, strconv.FormatFloat(in[k], 'f', -1, 64)); err != nil {
			return err
		}
	}
	if err := s.audit(ctx, tx, 0, "penjual", "ubah_pengaturan", before, after, "Parameter produksi diubah"); err != nil {
		return err
	}
	return tx.Commit()
}

// UpdateWeights menghitung bobot AHP dari matriks 3x3 (urutan: jam kirim,
// jumlah pesanan, waktu pesan). Hanya disimpan bila CR < 0,1; bila tidak,
// consistent = false dan handler membalas 422.
func (s Service) UpdateWeights(ctx context.Context, matrix [][]float64) (map[string]interface{}, bool, error) {
	weights, cr, err := AHP(matrix)
	if err != nil {
		return nil, false, errors.Join(bakery.ErrValidation, err)
	}
	resp := map[string]interface{}{"weights": weights, "cr": cr, "consistent": cr < 0.1}
	if cr >= 0.1 {
		resp["error"] = "penilaian belum konsisten (CR >= 0,1), mohon revisi"
		return resp, false, nil
	}
	raw, _ := json.Marshal(matrix)
	if err := s.SaveWeights(ctx, string(raw), weights, cr); err != nil {
		return nil, false, err
	}
	return resp, true, nil
}

// SaveWeights menyimpan bobot AHP sebagai versi baru (versi lama dinonaktifkan).
func (s Service) SaveWeights(ctx context.Context, matrixJSON string, w []float64, cr float64) error {
	if len(w) != 3 {
		return fmt.Errorf("%w: harus tepat 3 kriteria (jam kirim, jumlah, waktu pesan)", bakery.ErrValidation)
	}
	tx, err := s.bakery.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := s.bakery.DeactivateWeights(ctx, tx); err != nil {
		return err
	}
	if err := s.bakery.InsertWeights(ctx, tx, matrixJSON, w, cr); err != nil {
		return err
	}
	if err := s.audit(ctx, tx, 0, "penjual", "ubah_bobot", nil,
		map[string]interface{}{"J": w[0], "Q": w[1], "W": w[2], "CR": cr}, "Bobot AHP baru diaktifkan"); err != nil {
		return err
	}
	return tx.Commit()
}

func (s Service) AuditLog(ctx context.Context, limit int) ([]bakery.AuditEntry, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	return s.bakery.GetAuditLog(ctx, limit)
}

func (s Service) Notifications(ctx context.Context) ([]bakery.Notification, error) {
	return s.bakery.ListNotifications(ctx)
}

func (s Service) MarkNotificationsRead(ctx context.Context) error {
	return s.bakery.MarkNotificationsRead(ctx)
}
