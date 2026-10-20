-- Migrasi: pesanan besar (> 400 pcs) wajib dipesan paling lambat H-7 karena bahan
-- baku dibeli di luar kota. Pesanan besar yang lebih mendadak hanya diterima bila
-- penjual menyatakan stok bahan baku cukup. Aman dijalankan berulang.
INSERT INTO settings (kunci, nilai) VALUES
  ('batas_pesanan_besar','400'), ('min_hari_pesanan_besar','7')
  ON DUPLICATE KEY UPDATE nilai=VALUES(nilai), diubah_pada=CURRENT_TIMESTAMP;
