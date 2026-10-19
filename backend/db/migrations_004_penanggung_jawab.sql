-- Migrasi: penanggung jawab order (akun yang menginput pesanan).
-- Jalankan SEKALI pada database yang sudah ada (ALTER TABLE tidak idempotent).
-- Order lama tetap NULL = penanggung jawab belum tercatat.
ALTER TABLE orders
  ADD COLUMN created_by BIGINT NULL AFTER catatan,
  ADD CONSTRAINT fk_orders_user FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;
