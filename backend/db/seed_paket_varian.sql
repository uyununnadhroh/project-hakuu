-- Contoh data paket pilih varian (opsional, untuk demo/pengembangan).
-- Butuh migrations_006_paket_varian.sql (atau schema.sql terbaru). Harga & varian silakan disesuaikan.
-- Aman dijalankan ulang: paket/varian yang sudah ada dilewati.
INSERT INTO products (nama, harga, kategori, jenis, bobot, maks_varian)
SELECT n.nama, n.harga, 'Paket Pilih Varian', 'campuran', 1.10, n.maks
FROM (
  SELECT 'Paket Hemat isi 2' AS nama, 30000 AS harga, 2 AS maks
  UNION ALL SELECT 'Paket Hemat isi 3', 42000, 3
  UNION ALL SELECT 'Paket Hemat isi 4', 55000, 4
) n
WHERE NOT EXISTS (SELECT 1 FROM products p WHERE p.nama = n.nama);

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
WHERE p.kategori = 'Paket Pilih Varian' AND p.maks_varian > 0
  AND NOT EXISTS (SELECT 1 FROM product_variants pv WHERE pv.product_id = p.id AND pv.nama = v.nama);
