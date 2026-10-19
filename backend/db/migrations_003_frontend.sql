-- Migrasi untuk database yang dibuat dari schema.sql versi sebelumnya.
-- Jalankan SEKALI (ALTER TABLE tidak idempotent).
ALTER TABLE customers ADD COLUMN alamat VARCHAR(255) NULL AFTER telepon;

CREATE TABLE IF NOT EXISTS categories (
  id   BIGINT AUTO_INCREMENT PRIMARY KEY,
  nama VARCHAR(60) NOT NULL UNIQUE
) ENGINE=InnoDB;
INSERT IGNORE INTO categories (nama) VALUES ('Menu Hemat'), ('Menu Paket');

ALTER TABLE products
  ADD COLUMN kategori VARCHAR(60) NOT NULL DEFAULT 'Menu Hemat',
  ADD COLUMN jenis ENUM('tanpa_topping','bertopping','campuran') NOT NULL DEFAULT 'tanpa_topping',
  ADD COLUMN bobot DECIMAL(4,2) NOT NULL DEFAULT 1.00;

-- jumlah = beban kapasitas, qty_box = jumlah box. Order lama: keduanya sama.
ALTER TABLE orders ADD COLUMN qty_box INT NOT NULL DEFAULT 0 AFTER jumlah;
UPDATE orders SET qty_box = jumlah WHERE qty_box = 0;

-- Tanpa lembur: proofing roti tidak bisa dipercepat tanpa mengubah rasa/ketahanan.
-- Kelebihan pesanan dari H-1 dialihkan ke hari H (jam kirim >= 11.00) atau dinegosiasikan.
INSERT INTO settings (kunci, nilai) VALUES ('lembur_maks','0')
  ON DUPLICATE KEY UPDATE nilai='0', diubah_pada=CURRENT_TIMESTAMP;
UPDATE daily_capacity SET lembur_maks = 0;
