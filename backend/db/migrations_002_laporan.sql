-- Migrasi untuk database yang sudah dibuat dari schema.sql versi awal.
CREATE TABLE IF NOT EXISTS monthly_reports (
  bulan       CHAR(7) PRIMARY KEY,          -- contoh: 2026-10
  data        JSON NOT NULL,
  dibuat_pada DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;
INSERT IGNORE INTO settings (kunci, nilai) VALUES ('jam_peringatan_sebelum_selesai','2');
