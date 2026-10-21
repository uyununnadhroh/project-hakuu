# bakery-scheduler (Backend Penjadwalan Produksi, Go + MySQL)

Struktur folder mengikuti pola **panakea-be**: `cmd → boot → delivery (handler) → service → data`,
dengan `entity` untuk struct dan `config` + file yaml untuk koneksi database.

## Struktur

```
bakery-scheduler/
├── cmd/
│   ├── http/main.go              # entry point server  (go run ./cmd/http)
│   └── createuser/main.go        # buat akun penjual    (go run ./cmd/createuser ...)
├── files/etc/bakery-scheduler/
│   └── bakery-scheduler.development.yaml   # port, DSN database, JWT secret, CORS
├── db/                           # schema.sql + migrasi (tidak berubah)
├── internal/
│   ├── boot/
│   │   ├── http.go               # baca config, buka DB, rangkai data → service → handler, start server
│   │   └── cron.go               # job per jam (RunHourly)
│   ├── config/                   # config.go, config_key.go (struct yaml)
│   ├── entity/
│   │   ├── context.go
│   │   └── bakery/               # SEMUA struct: request, response, laporan, + error standar
│   ├── data/bakery/              # SEMUA query SQL
│   │   ├── init.go               # Data struct, daftar readStmt/insertStmt/updateStmt/deleteStmt, InitStmt
│   │   ├── init_*.go             # konstanta query (init_order.go, init_master.go, ...)
│   │   └── *.go                  # fungsi yang menjalankan query (order.go, master.go, ...)
│   ├── service/bakery/           # logika bisnis (tidak ada SQL di sini)
│   │   ├── init.go               # interface BakeryData + struct Service + New
│   │   ├── order.go              # terima order, geser, lembur, negosiasi jam, tolak
│   │   ├── order_ops.go          # start/complete/cancel, job per jam, rebalance
│   │   ├── order_query.go        # daftar order, detail, jadwal
│   │   ├── negosiasi.go, pengaturan.go, master.go, laporan.go, auth.go
│   │   ├── rules.go              # aturan murni: skor, deadline, AHP, jam alternatif
│   │   └── helper.go             # settings, kunci kapasitas, audit, label status
│   └── delivery/http/
│       ├── server.go             # interface BakeryHandler + struct Server + Serve
│       ├── handler.go            # DAFTAR ROUTE (mux)
│       ├── jwt_middleware.go, cors_middleware.go
│       └── bakery/               # handler per fitur (order.go, master.go, ...)
└── pkg/
    ├── response/response.go      # Response{} + RenderJSON
    └── grace/grace.go            # graceful shutdown (sama dengan panakea)
```

### Alur satu request (contoh `GET /api/products`)
`delivery/http/handler.go` (route) → `delivery/http/bakery/master.go` `ListProducts` (baca request)
→ `service/bakery/master.go` `ListProducts` (aturan bisnis) → `data/bakery/master.go` `ListProducts`
(jalankan query `qListProducts` dari `init_master.go`).

### Peta dari struktur lama
| Lama | Baru |
|---|---|
| `cmd/server/main.go` | `cmd/http/main.go` + `internal/boot/` |
| env `DB_DSN`, `JWT_SECRET`, `CORS_ORIGINS` | `files/etc/bakery-scheduler/bakery-scheduler.development.yaml` (env masih bisa menimpa) |
| `internal/api/server.go` | `internal/delivery/http/` (routes, middleware) + `internal/delivery/http/bakery/` (handler) |
| `internal/store/*.go` (SQL + logika campur) | SQL → `internal/data/bakery/`, logika → `internal/service/bakery/` |
| `internal/domain/` | `internal/service/bakery/rules.go`, tipe-tipenya di `internal/entity/bakery/rules.go` |
| struct di `store` (OrderRow, Report, ...) | `internal/entity/bakery/` |

**Yang tidak berubah:** URL endpoint, method, format JSON request/response, dan kode status HTTP.
Frontend `hakuu-web` tidak perlu diubah. Database juga tidak berubah.

## Menjalankan (Windows PowerShell + XAMPP)
1. Start MySQL di XAMPP. Database `bakery` sudah ada (dari `db/schema.sql`).
   Database lama: jalankan juga `db/migrations_002_laporan.sql`, `db/migrations_003_frontend.sql`, `db/migrations_004_penanggung_jawab.sql`, `db/migrations_005_jam_kirim.sql`, `db/migrations_006_paket_varian.sql`, `db/migrations_007_konfirmasi_pengiriman.sql`, `db/migrations_008_jam_operasional.sql`, `db/migrations_009_pesanan_besar.sql`, dan `db/migrations_010_isi_box.sql` (berurutan; 002–004, 006, dan 010 cukup sekali).
   Contoh paket pilih varian (Paket Hemat isi 2/3/4): `db/seed_paket_varian.sql` (opsional).
2. Cek `files/etc/bakery-scheduler/bakery-scheduler.development.yaml`: DSN database dan ganti `jwt.secret`
   dengan string acak minimal 32 karakter.
3. Dari folder project:
   ```powershell
   go mod tidy
   go run ./cmd/createuser -email penjual@contoh.com -nama "Nama" -password "minimal8karakter"
   go run ./cmd/http
   ```
   Server berjalan di `http://localhost:8080`. Cek: buka `http://localhost:8080/api/BEversion`.

Saat start, semua query di-*prepare* dulu (pola panakea). Bila muncul
`Failed to initialize ... statement key ...` artinya ada tabel/kolom yang belum ada di database,
biasanya karena migrasi di langkah 1 belum dijalankan.

## Aturan hari produksi

Jam operasional produksi **06.00–18.00** (`jam_produksi_mulai`, `jam_produksi_selesai`), jam kirim
paling lambat **20.00** (`jam_kirim_terakhir`; lebih dari itu order ditolak). Tidak ada produksi H-2.
Saat order masuk, sistem memeriksa:

0. **Pesanan besar mendadak**: lebih dari 400 pcs (`batas_pesanan_besar`) dan dipesan kurang dari
   7 hari sebelum tanggal kirim (`min_hari_pesanan_besar`). Bahan baku dibeli di luar kota, jadi
   sistem belum menjadwalkan apa pun: hasilnya `perlu_cek_bahan` dan penjual ditanya apakah stok
   bahan baku cukup. Order dikirim ulang dengan `"bahan_siap": true` (lanjut ke langkah 1) atau
   `"bahan_siap": false` (ditolak dengan alasan bahan baku).
1. Jam kirim **< 11.00** (`batas_pagi_jam`) → wajib diproduksi **H-1**.
2. Jam kirim **≥ 11.00** → diproduksi **H-1** bila kapasitas masih cukup; bila tidak, produksi
   **dialihkan ke hari H (hari kirim)** bila kapasitas hari H cukup.
3. Masih tidak muat → sistem mencoba menggeser order lain yang skor prioritasnya lebih rendah
   (hanya ke H-1/hari H milik order itu sendiri). Order yang sudah diterima tidak pernah dibatalkan
   atau digeser sampai terlambat. Lembur tidak dipakai (`lembur_maks` = 0).
4. Masih tidak muat → hasil `perlu_negosiasi` dengan rekomendasi jam kirim **di tanggal yang sama**,
   mulai 11.00 (produksi hari H) sampai paling lambat 20.00, maksimal `maks_tawaran_jam` pilihan.
5. Tidak ada jam yang muat sampai 20.00 → ditolak. Tidak ada saran pindah tanggal.

Produksi hanya bisa dimulai pada hari produksinya, mulai pukul 06.00. Order yang masuk setelah 18.00
tidak bisa lagi dijadwalkan untuk produksi hari itu. Order lama yang sudah terjadwal di H-2 tidak
dipindahkan otomatis.

## Endpoint (semua kecuali login butuh header `Authorization: Bearer <token>`)
| Method | Path | Fungsi |
|---|---|---|
| POST | /api/login | Login, mengembalikan token |
| GET | /api/users | Daftar email/akun aktif yang bisa masuk (tanpa password) |
| POST | /api/orders | Terima order: diterima / perlu_negosiasi / perlu_cek_bahan / ditolak (body opsional `"bahan_siap": true/false`) |
| GET | /api/orders?limit=1000, /api/orders/{id} | Daftar (maks. 1000) dan detail, termasuk `created_by_nama` & `created_by_email` (penanggung jawab) |
| POST | /api/orders/{id}/start, /complete, /cancel | Ubah status |
| GET | /api/schedule?from=YYYY-MM-DD&days=7 | Kapasitas dan order per hari |
| POST | /api/offers/{id}/offer | Ajukan jam alternatif (kapasitas ditahan) |
| POST | /api/offers/{id}/respond `{"accept":true}` | Catat jawaban pembeli |
| GET / PUT | /api/settings | Parameter produksi |
| PUT | /api/settings/weights | Bobot AHP (ditolak bila CR >= 0,1) |
| GET | /api/reports/monthly?month=2026-10 | Laporan bulanan |
| GET | /api/reports/monthly/export?month=2026-10 | Unduh CSV |
| POST | /api/reports/monthly/2026-10/close | Tutup bulan |
| GET | /api/audit-log, /api/notifications | Jejak perubahan dan notifikasi |
| POST | /api/notifications/read | Tandai notifikasi dibaca |
| GET | /api/delivery-confirmations | Pesanan yang perlu dikonfirmasi pengirimannya (batch yang sudah dibuka) |
| POST | /api/delivery-confirmations/{orderId} | Konfirmasi pengiriman: `{"status":"selesai"}` atau `{"status":"belum_terkirim","alasan":"..."}` |
| GET/POST | /api/products · PATCH/DELETE /api/products/{id} | Produk |
| GET/POST | /api/categories · DELETE /api/categories/{nama} | Kategori |
| GET/POST | /api/customers | Pelanggan |

Contoh order: `{"items":[{"product_id":1,"qty":30}],"deliver_at":"2026-10-20 15:00","customer":{"nama":"Bu Sari","telepon":"0812","alamat":"Jl. Mawar"}}`

Bobot kapasitas produk dihitung otomatis dari `jenis`, `isi_box` (roti per box, default 3), dan `isi_topping`
(roti bertopping, wajib untuk campuran). Patokannya box isi 3 bertopping = 1,23, jadi 1 roti bertopping = 0,41:
tanpa topping = 1,00 per box; bertopping = 0,41 × isi; campuran = 0,41 × roti bertopping + roti tanpa topping ÷ isi.
Contoh `{"jenis":"campuran","isi_box":4,"isi_topping":2}` → bobot 1,32. Kirim `bobot` hanya untuk mengisi manual.

Paket pilih varian lewat `POST /api/products` / `PATCH /api/products/{id}`: kirim `maks_varian` (isi paket, 0 = produk biasa)
dan `variants` (daftar lengkap, menggantikan yang lama). Contoh:
`{"maks_varian":3,"variants":[{"id":2,"nama":"Coklat"},{"nama":"Lava","opsi_label":"Isian selai","opsi_pilihan":["Stroberi","Nanas"]},{"nama":"Keju","aktif":false}]}`.
Varian aktif minimal sebanyak `maks_varian`. Varian lama yang tidak dikirim dihapus, atau dinonaktifkan bila sudah pernah dipesan.
`GET /api/products` mengirim semua varian beserta `aktif`; Input Pesanan hanya menampilkan yang aktif.

Konfirmasi pengiriman: batch pagi dibuka 11.00 (kirim hari ini sebelum 11.00), batch malam 20.00
(kirim hari ini mulai 11.00), ditambah pesanan 7 hari terakhir yang belum dikonfirmasi/belum terkirim.
"Selesai" menutup pesanan dan tidak bisa diubah lagi; "Belum terkirim" (alasan opsional) masih bisa
diubah ke "Selesai". Setiap konfirmasi disimpan di `delivery_confirmations` (status, waktu, user, alasan).
Notifikasi (`pending`) tetap aktif selama ada pesanan di batch terbuka yang belum Selesai. Semua jam
konfirmasi pengiriman dihitung dalam WIB (Asia/Jakarta), tidak terpengaruh `APP_TZ`.

## Menambah fitur baru (pola yang sama)
1. Struct request/response → `internal/entity/bakery/bakery.go`
2. Query → konstanta di `internal/data/bakery/init_xxx.go`, daftarkan di `init.go` (readStmt/insertStmt/...),
   fungsinya di `internal/data/bakery/xxx.go`
3. Tambah fungsi itu ke interface `BakeryData` di `internal/service/bakery/init.go`, logikanya di `service/bakery/xxx.go`
4. Tambah ke interface `Service` di `internal/delivery/http/bakery/init.go`, handlernya di `delivery/http/bakery/xxx.go`
5. Tambah ke interface `BakeryHandler` di `internal/delivery/http/server.go`, daftarkan route di `handler.go`

## Pengujian
- `go test ./internal/service/bakery/ ./internal/delivery/http/` : aturan murni + CORS (tanpa database).
- Tes integrasi butuh database KOSONG berisi schema.sql (tes ini MENGHAPUS isi tabel, jangan ke data asli):
  `$env:TEST_DB_DSN="root:@tcp(localhost:3306)/bakery_test?parseTime=true&loc=Asia%2FJakarta"; go test ./internal/service/bakery/`
