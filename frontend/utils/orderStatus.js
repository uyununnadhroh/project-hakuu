// ─── Helper tampilan pesanan ──────────────────────────────────────────────
// Status pesanan dihitung oleh SERVER (backend). Di sini hanya memetakan
// status backend ke label yang tampil di UI, plus helper tanggal.

const pad = (n) => String(n).padStart(2, '0');

/** Tanggal lokal browser -> 'YYYY-MM-DD' (bukan UTC, supaya tidak mundur sehari di WIB). */
export function ymd(d = new Date()) {
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
}

export function isoDate(offsetDays = 0) {
  const d = new Date();
  d.setDate(d.getDate() + offsetDays);
  return ymd(d);
}

function parseYmd(iso) {
  const [y, m, d] = iso.slice(0, 10).split('-').map(Number);
  return new Date(y, m - 1, d);
}

export function fmtTgl(iso) {
  if (!iso) return '—';
  return parseYmd(iso).toLocaleDateString('id-ID', { day: 'numeric', month: 'short' });
}

/** Selisih hari a - b (keduanya 'YYYY-MM-DD'). */
export function dayDiff(a, b) {
  const t = (s) => { const d = parseYmd(s); return Date.UTC(d.getFullYear(), d.getMonth(), d.getDate()); };
  return Math.round((t(a) - t(b)) / 86400000);
}

/** Hari produksi relatif terhadap hari kirim: 'H' (hari itu juga) atau 'H-1', 'H-2', ... */
export function hLabel(tanggalKirim, tanggalProduksi) {
  if (!tanggalProduksi) return '—';
  const n = dayDiff(tanggalKirim, tanggalProduksi);
  return n <= 0 ? 'H' : `H-${n}`;
}

/**
 * Status backend -> label UI.
 *   baru / menunggu_negosiasi -> Diterima   (di luar jendela 72 jam)
 *   terjadwal / at_risk       -> Dijadwalkan  (at_risk tidak diberi badge sendiri)
 *   diproduksi                -> Diproduksi   (setelah pesanan dicetak / "Mulai produksi")
 *   terlambat                 -> Terlambat    (lewat jam kirim, belum diselesaikan)
 *   selesai / batal           -> Selesai / Dibatalkan
 */
export function uiStatus(o) {
  switch (o.status) {
    case 'batal':      return 'Dibatalkan';
    case 'selesai':    return 'Selesai';
    case 'terlambat':  return 'Terlambat';
    case 'diproduksi': return 'Diproduksi';
    case 'terjadwal':
    case 'at_risk':    return 'Dijadwalkan';
    default:           return 'Diterima';
  }
}

/** Teks badge status: Dijadwalkan diberi tanggal produksinya, mis. "Dijadwalkan tgl 20 Okt". */
export function statusText(o) {
  if (o.statusOtomatis === 'Dijadwalkan' && o.tanggalProduksi) return `Dijadwalkan tgl ${fmtTgl(o.tanggalProduksi)}`;
  return o.statusOtomatis;
}

const base = 'inline-flex items-center px-2.5 py-1 rounded-lg text-xs font-medium whitespace-nowrap';
export const badgeMap = {
  'Dijadwalkan': `${base} border border-[#4c0c13] text-[#4c0c13] bg-white`,
  'Diterima':    `${base} bg-[#fcebc8] text-[#7e4722]`,
  'Diproduksi':  `${base} bg-[#fff0e0] text-[#c2410c]`,
  'Selesai':     `${base} bg-[#dcfce7] text-[#166534]`,
  'Dibatalkan':  `${base} bg-[#fee2e2] text-[#991b1b]`,
  'Terlambat':   `${base} bg-[#fee2e2] text-[#b74b40] border border-[#f3b8b2]`,
};
