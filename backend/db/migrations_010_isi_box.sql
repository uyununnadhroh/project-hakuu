-- Migrasi: bobot kapasitas dihitung dari jenis dan isi box.
-- Patokan: 1 box isi 3 roti bertopping = 1,23, jadi 1 roti bertopping = 0,41.
--   tanpa topping : 1,00 per box (isi berapa pun)
--   bertopping    : 0,41 x isi box
--   campuran      : 0,41 x roti bertopping + roti tanpa topping / isi box
-- Jalankan SEKALI (ALTER TABLE tidak idempotent). Produk lama dianggap box isi 3;
-- campuran dianggap 1 roti bertopping + 2 tanpa topping. Sesuaikan isi box tiap
-- produk lewat halaman Kelola Produk, bobotnya akan dihitung ulang otomatis.
ALTER TABLE products
  ADD COLUMN isi_box     TINYINT NOT NULL DEFAULT 3 AFTER bobot,
  ADD COLUMN isi_topping TINYINT NOT NULL DEFAULT 0 AFTER isi_box;

UPDATE products SET isi_topping = isi_box WHERE jenis = 'bertopping';
UPDATE products SET isi_topping = 1       WHERE jenis = 'campuran';
UPDATE products SET bobot = CASE jenis
    WHEN 'bertopping' THEN ROUND(0.41 * isi_box, 2)
    WHEN 'campuran'   THEN ROUND(0.41 * isi_topping + (isi_box - isi_topping) / isi_box, 2)
    ELSE 1.00
  END;
