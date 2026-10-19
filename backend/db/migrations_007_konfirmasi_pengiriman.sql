-- Migrasi: konfirmasi pengiriman per pesanan (batch pagi 11.00 & malam 20.00).
-- Aman dijalankan ulang.
CREATE TABLE IF NOT EXISTS delivery_confirmations (
  id                BIGINT AUTO_INCREMENT PRIMARY KEY,
  order_id          BIGINT NOT NULL,
  status            ENUM('selesai','belum_terkirim') NOT NULL,
  alasan            VARCHAR(255) NULL,
  dikonfirmasi_oleh BIGINT NULL,
  dikonfirmasi_pada DATETIME NOT NULL,
  INDEX idx_dc_order (order_id, id),
  FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
  FOREIGN KEY (dikonfirmasi_oleh) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB;
