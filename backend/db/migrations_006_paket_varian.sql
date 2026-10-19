-- Migrasi: paket pilih varian (Paket Hemat isi 2/3/4, dst).
-- Jalankan SEKALI (ALTER TABLE tidak idempotent). Contoh datanya ada di seed_paket_varian.sql.
ALTER TABLE products
  ADD COLUMN maks_varian TINYINT NOT NULL DEFAULT 0 AFTER bobot;

CREATE TABLE IF NOT EXISTS product_variants (
  id           BIGINT AUTO_INCREMENT PRIMARY KEY,
  product_id   BIGINT NOT NULL,
  nama         VARCHAR(60) NOT NULL,
  opsi_label   VARCHAR(60) NULL,
  opsi_pilihan JSON NULL,
  aktif        BOOLEAN NOT NULL DEFAULT TRUE,
  urutan       INT NOT NULL DEFAULT 0,
  UNIQUE KEY uq_product_variant (product_id, nama),
  FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS order_item_variants (
  id            BIGINT AUTO_INCREMENT PRIMARY KEY,
  order_item_id BIGINT NOT NULL,
  variant_id    BIGINT NOT NULL,
  opsi          VARCHAR(60) NULL,
  FOREIGN KEY (order_item_id) REFERENCES order_items(id) ON DELETE CASCADE,
  FOREIGN KEY (variant_id) REFERENCES product_variants(id)
) ENGINE=InnoDB;

INSERT IGNORE INTO categories (nama) VALUES ('Paket Pilih Varian');
