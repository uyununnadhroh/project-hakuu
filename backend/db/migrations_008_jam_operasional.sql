-- Migrasi: jam operasional produksi 06.00–18.00, jam kirim paling lambat 20.00,
-- dan batas H-1 jam 11.00 (jam kirim < 11.00 wajib diproduksi H-1).
-- Batas H-1 kini tidak lagi dihitung dari jam produksi selesai + packing + tempuh.
-- Aman dijalankan berulang.
INSERT INTO settings (kunci, nilai) VALUES
  ('jam_produksi_mulai','6'), ('jam_produksi_selesai','18'),
  ('jam_kirim_terakhir','20'), ('batas_pagi_jam','11')
  ON DUPLICATE KEY UPDATE nilai=VALUES(nilai), diubah_pada=CURRENT_TIMESTAMP;
DELETE FROM settings WHERE kunci IN ('menit_packing','menit_tempuh');
