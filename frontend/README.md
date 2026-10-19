# Hakuu Web (Next.js)

Frontend Hakuu Bakery — struktur mengikuti pola proyek Panakea (Next.js, pages router).

## Menjalankan

```bash
npm install
npm run dev        # http://localhost:5173
```

Alamat backend diatur di `.env.development` (`NEXT_PUBLIC_BASE_API_URL`).

## Struktur folder

```
components/          komponen yang dipakai bersama
  Routes.js          daftar menu sidebar + judul di navbar
  Sidebar.js         sidebar kiri (menu + tombol Keluar)
  Navbar.js          bar atas (Hakuu Bakery / nama halaman)
  ErrorBanner.js     kotak pesan error merah
  KonfirmasiPengiriman.js  tombol + jendela konfirmasi pengiriman di Navbar (batch 11.00 & 20.00)
  pesanan/           komponen halaman Input Pesanan
    PackageCard.js   kartu paket pilih varian (qty + pilih varian)
    QtyStepper.js    stepper jumlah box (−, angka, +)
    VariantPicker.js chip varian, batas maksimum, opsi tambahan (mis. isian selai)
  produk/            komponen halaman Kelola Produk
    VariantEditor.js isi paket + daftar varian (tambah, hapus, aktif/nonaktif, pilihan tambahan)

pages/               setiap file = satu halaman (URL)
  _app.js            kerangka aplikasi: cek login, pasang Sidebar + Navbar
  _document.js       <html lang="id">, favicon
  index.js           "/"  -> redirect ke /dashboard atau /login
  404.js             URL tidak dikenal -> redirect ke "/"
  login.js           /login
  dashboard/index.js         /dashboard
  input-pesanan/index.js     /input-pesanan
  daftar-pesanan/index.js    /daftar-pesanan
  jadwal-produksi/index.js   /jadwal-produksi
  laporan-penjualan/index.js /laporan-penjualan  (rekap bulanan & cetak laporan)
  kelola-produk/index.js     /kelola-produk
  pengaturan/index.js        /pengaturan

services/            semua pemanggilan API backend
  core.js            login
  bakery.js          pesanan, penawaran, jadwal, pengaturan, produk, kategori

redux/               state global (nama user yang login)
  store.js, actions/userAction.js, reducers/

utils/               helper
  apiClient.js       instance axios (token, error handling, auto-logout 401)
  storage.js         simpan/ambil token & nama di localStorage/sessionStorage
  mapper.js          ubah data backend -> bentuk yang dipakai tampilan
  orderStatus.js     label & warna status pesanan, helper tanggal
  useLoad.js         hook untuk load data + loading + error
  icons.js           URL ikon & logo

styles/globals.css   Tailwind + class custom (btn-primary, card, input-field, ...)
public/              favicon
```

## Menambah halaman baru

1. Buat `pages/nama-halaman/index.js`.
2. Tambahkan fungsi API-nya di `services/bakery.js`.
3. Tambahkan menunya di `components/Routes.js`.
