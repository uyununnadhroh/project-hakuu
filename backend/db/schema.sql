-- MySQL 8 / MariaDB 10.6+
SET NAMES utf8mb4;

CREATE TABLE users (
  id            BIGINT AUTO_INCREMENT PRIMARY KEY,
  nama          VARCHAR(100) NOT NULL,
  email         VARCHAR(150) NOT NULL UNIQUE,
  password_hash VARCHAR(100) NOT NULL,
  aktif         BOOLEAN NOT NULL DEFAULT TRUE,
  dibuat_pada   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;

CREATE TABLE customers (
  id      BIGINT AUTO_INCREMENT PRIMARY KEY,
  nama    VARCHAR(120) NOT NULL,
  telepon VARCHAR(30) NULL,
  alamat  VARCHAR(255) NULL,
  catatan VARCHAR(255) NULL
) ENGINE=InnoDB;

CREATE TABLE categories (
  id   BIGINT AUTO_INCREMENT PRIMARY KEY,
  nama VARCHAR(60) NOT NULL UNIQUE
) ENGINE=InnoDB;

CREATE TABLE products (
  id       BIGINT AUTO_INCREMENT PRIMARY KEY,
  nama     VARCHAR(120) NOT NULL,
  harga    DECIMAL(12,2) NOT NULL,
  aktif    BOOLEAN NOT NULL DEFAULT TRUE,
  kategori VARCHAR(60) NOT NULL DEFAULT 'Menu Hemat',
  jenis    ENUM('tanpa_topping','bertopping','campuran') NOT NULL DEFAULT 'tanpa_topping',
  bobot    DECIMAL(4,2) NOT NULL DEFAULT 1.00,  -- bobot kapasitas per box (dihitung dari jenis + isi box)
  isi_box     TINYINT NOT NULL DEFAULT 3,        -- jumlah roti per box (patokan bobot: box isi 3)
  isi_topping TINYINT NOT NULL DEFAULT 0,        -- jumlah roti bertopping di dalam box (campuran)
  maks_varian TINYINT NOT NULL DEFAULT 0       -- paket pilih varian: jumlah varian wajib dipilih (0 = bukan paket varian)
) ENGINE=InnoDB;

-- Varian yang bisa dipilih untuk paket (products.maks_varian > 0).
-- opsi_label/opsi_pilihan: pilihan tambahan bila varian dipilih, mis. Lava -> 'Isian selai' ["Stroberi","Blueberry"].
CREATE TABLE product_variants (
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

CREATE TABLE orders (
  id                 BIGINT AUTO_INCREMENT PRIMARY KEY,
  kode               VARCHAR(20) NULL UNIQUE,
  customer_id        BIGINT NULL,
  jumlah             INT NOT NULL,           -- beban kapasitas (box x bobot produk)
  qty_box            INT NOT NULL DEFAULT 0, -- jumlah box yang dipesan
  total_harga        DECIMAL(14,2) NOT NULL DEFAULT 0,
  waktu_pesan        DATETIME NOT NULL,
  waktu_kirim        DATETIME NOT NULL,
  waktu_kirim_awal   DATETIME NOT NULL,
  hari_deadline      DATE NOT NULL,
  hari_produksi      DATE NULL,
  lembur_pcs         INT NOT NULL DEFAULT 0,
  status             ENUM('baru','terjadwal','menunggu_negosiasi','diproduksi',
                          'selesai','at_risk','terlambat','batal') NOT NULL DEFAULT 'baru',
  skor_j DOUBLE NULL, skor_q DOUBLE NULL, skor_w DOUBLE NULL, skor_total DOUBLE NULL,
  skor_dihitung_pada DATETIME NULL,
  diserahkan_pada    DATETIME NULL,
  catatan            VARCHAR(255) NULL,
  created_by         BIGINT NULL,            -- user yang menginput order (penanggung jawab)
  CONSTRAINT chk_jumlah CHECK (jumlah > 0),
  CONSTRAINT fk_orders_customer FOREIGN KEY (customer_id) REFERENCES customers(id),
  CONSTRAINT fk_orders_user FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
  INDEX idx_status_kirim (status, waktu_kirim)
) ENGINE=InnoDB;

CREATE TABLE order_items (
  id           BIGINT AUTO_INCREMENT PRIMARY KEY,
  order_id     BIGINT NOT NULL,
  product_id   BIGINT NOT NULL,
  jumlah       INT NOT NULL,
  harga_satuan DECIMAL(12,2) NOT NULL,
  subtotal     DECIMAL(14,2) NOT NULL,
  FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
  FOREIGN KEY (product_id) REFERENCES products(id)
) ENGINE=InnoDB;

-- Varian yang dipilih per baris paket; berlaku untuk semua box di baris itu.
CREATE TABLE order_item_variants (
  id            BIGINT AUTO_INCREMENT PRIMARY KEY,
  order_item_id BIGINT NOT NULL,
  variant_id    BIGINT NOT NULL,
  opsi          VARCHAR(60) NULL,
  FOREIGN KEY (order_item_id) REFERENCES order_items(id) ON DELETE CASCADE,
  FOREIGN KEY (variant_id) REFERENCES product_variants(id)
) ENGINE=InnoDB;

CREATE TABLE daily_capacity (
  tanggal         DATE PRIMARY KEY,
  kapasitas_dasar INT NOT NULL,
  lembur_maks     INT NOT NULL,
  terkunci        BOOLEAN NOT NULL DEFAULT FALSE
) ENGINE=InnoDB;

CREATE TABLE capacity_reservations (
  id               BIGINT AUTO_INCREMENT PRIMARY KEY,
  order_id         BIGINT NOT NULL,
  tanggal          DATE NOT NULL,
  jumlah           INT NOT NULL,
  lembur           INT NOT NULL DEFAULT 0,
  tipe             ENUM('tetap','tahanan_sementara') NOT NULL DEFAULT 'tetap',
  kedaluwarsa_pada DATETIME NULL,
  FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
  FOREIGN KEY (tanggal) REFERENCES daily_capacity(tanggal),
  INDEX idx_res_tanggal (tanggal)
) ENGINE=InnoDB;

CREATE TABLE negotiation_offers (
  id             BIGINT AUTO_INCREMENT PRIMARY KEY,
  order_id       BIGINT NOT NULL,
  urutan         INT NOT NULL,
  jam_alternatif DATETIME NOT NULL,
  hari_produksi  DATE NOT NULL,
  status         ENUM('direkomendasikan','ditawarkan','diterima','ditolak','kedaluwarsa')
                 NOT NULL DEFAULT 'direkomendasikan',
  ditawarkan_pada DATETIME NULL,
  batas_respons   DATETIME NULL,
  reservation_id  BIGINT NULL,
  dijawab_pada    DATETIME NULL,
  FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
) ENGINE=InnoDB;

CREATE TABLE settings (
  kunci       VARCHAR(50) PRIMARY KEY,
  nilai       VARCHAR(50) NOT NULL,
  diubah_pada DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;

CREATE TABLE ahp_weights (
  id          BIGINT AUTO_INCREMENT PRIMARY KEY,
  matriks     JSON NOT NULL,
  bobot_j DOUBLE NOT NULL, bobot_q DOUBLE NOT NULL, bobot_w DOUBLE NOT NULL,
  cr          DOUBLE NOT NULL,
  aktif       BOOLEAN NOT NULL DEFAULT FALSE,
  dibuat_pada DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;

CREATE TABLE audit_log (
  id           BIGINT AUTO_INCREMENT PRIMARY KEY,
  waktu        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  order_id     BIGINT NULL,
  aksi         VARCHAR(40) NOT NULL,
  data_sebelum JSON NULL,
  data_sesudah JSON NULL,
  alasan       VARCHAR(255) NULL,
  pelaku       VARCHAR(50) NOT NULL DEFAULT 'sistem',
  INDEX idx_audit_order (order_id),
  INDEX idx_audit_waktu (waktu)
) ENGINE=InnoDB;

-- Konfirmasi pengiriman oleh penjual (batch pagi 11.00 & malam 20.00).
-- Satu baris per konfirmasi; konfirmasi terakhir = status pengiriman saat ini.
CREATE TABLE delivery_confirmations (
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

CREATE TABLE notifications (
  id          BIGINT AUTO_INCREMENT PRIMARY KEY,
  order_id    BIGINT NULL,
  pesan       VARCHAR(255) NOT NULL,
  dibaca      BOOLEAN NOT NULL DEFAULT FALSE,
  dibuat_pada DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;

INSERT INTO settings (kunci, nilai) VALUES
  ('kapasitas_harian','800'),      ('lembur_maks','0'),
  ('jendela_jam','72'),            ('batas_pagi_jam','11'),
  ('lama_menunggu_maks_jam','96'), ('pembulatan_jam','1'),
  ('ambang_selisih_skor','0.05'),  ('maks_tawaran_jam','3'),
  ('batas_respons_menit','60'),    ('jam_produksi_mulai','6'),
  ('jam_produksi_selesai','18'),   ('jam_kirim_terakhir','20'),
  ('batas_pesanan_besar','400'),   ('min_hari_pesanan_besar','7');

INSERT INTO ahp_weights (matriks, bobot_j, bobot_q, bobot_w, cr, aktif)
VALUES ('[[1,3,5],[0.3333333333,1,2],[0.2,0.5,1]]', 0.648, 0.230, 0.122, 0.003, TRUE);
-- Migrasi untuk database yang sudah dibuat dari schema.sql versi awal.
CREATE TABLE IF NOT EXISTS monthly_reports (
  bulan       CHAR(7) PRIMARY KEY,          -- contoh: 2026-10
  data        JSON NOT NULL,
  dibuat_pada DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;
INSERT IGNORE INTO settings (kunci, nilai) VALUES ('jam_peringatan_sebelum_selesai','2');

INSERT INTO categories (nama) VALUES ('Menu Hemat'), ('Menu Paket'), ('Paket Pilih Varian');
