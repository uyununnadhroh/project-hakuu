-- Migrasi: rekomendasi jam kirim hanya di hari yang sama, mulai jam siap (11.00) sampai 19.00.
-- Aman dijalankan berulang.
INSERT INTO settings (kunci, nilai) VALUES ('jam_kirim_terakhir','19')
  ON DUPLICATE KEY UPDATE nilai='19', diubah_pada=CURRENT_TIMESTAMP;
