-- Migrasi: rencana produksi suatu hari kini dikunci mulai jam 08.00 di hari
-- produksi itu sendiri (sebelumnya 18.00 sehari sebelumnya). Hari yang terlanjur
-- dikunci dengan aturan lama dibuka lagi. Aman dijalankan berulang.
INSERT INTO settings (kunci, nilai) VALUES ('jam_kunci_rencana','8')
  ON DUPLICATE KEY UPDATE nilai=VALUES(nilai), diubah_pada=CURRENT_TIMESTAMP;
UPDATE daily_capacity SET terkunci = 0 WHERE tanggal > CURDATE();
