-- Contoh data paket pilih varian (opsional, untuk demo/pengembangan).
-- Butuh migrations_006_paket_varian.sql (atau schema.sql terbaru). Harga & varian silakan disesuaikan.
INSERT INTO products (nama, harga, kategori, jenis, bobot, maks_varian) VALUES
  ('Paket Hemat isi 2', 30000, 'Paket Pilih Varian', 'campuran', 1.10, 2),
  ('Paket Hemat isi 3', 42000, 'Paket Pilih Varian', 'campuran', 1.10, 3),
  ('Paket Hemat isi 4', 55000, 'Paket Pilih Varian', 'campuran', 1.10, 4);

-- Varian yang sama untuk ketiga paket di atas.
INSERT INTO product_variants (product_id, nama, opsi_label, opsi_pilihan, urutan)
SELECT p.id, v.nama, v.opsi_label, v.opsi_pilihan, v.urutan
FROM products p
JOIN (
  SELECT 'Coklat' AS nama, NULL AS opsi_label, NULL AS opsi_pilihan, 1 AS urutan
  UNION ALL SELECT 'Keju',  NULL, NULL, 2
  UNION ALL SELECT 'Abon',  NULL, NULL, 3
  UNION ALL SELECT 'Meses', NULL, NULL, 4
  UNION ALL SELECT 'Lava',  'Isian selai', JSON_ARRAY('Stroberi','Blueberry','Nanas'), 5
) v
WHERE p.kategori = 'Paket Pilih Varian' AND p.maks_varian > 0;
