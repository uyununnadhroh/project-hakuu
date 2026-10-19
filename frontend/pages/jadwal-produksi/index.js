import { useState } from 'react';
import api from '../../services/bakery';
import { useLoad } from '../../utils/useLoad';
import ErrorBanner from '../../components/ErrorBanner';
import { fmtTgl, ymd } from '../../utils/orderStatus';
import { penanggungJawab } from '../../utils/mapper';

function IconPrinter() {
  return (
    <svg viewBox="0 0 24 24" className="size-4 shrink-0" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M6 9V2h12v7" />
      <path d="M6 18H4a2 2 0 0 1-2-2v-5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v5a2 2 0 0 1-2 2h-2" />
      <rect x="6" y="14" width="12" height="8" />
    </svg>
  );
}

function ScoreBadge({ score }) {
  if (score == null) return <span className="text-[#8a7768] text-xs">—</span>;
  const s = Math.round(score * 100);
  const color = s >= 85 ? 'bg-[#eaf5ed] text-[#327a52]'
              : s >= 70 ? 'bg-[#fdf6ec] text-[#a66a17]'
              :           'bg-[#fef2f2] text-[#b74b40]';
  return <span className={`inline-flex items-center px-2 py-0.5 rounded-lg text-xs font-medium ${color}`}>{s}</span>;
}

function SesiTable({ title, subtitle, rows }) {
  return (
    <div className="card">
      <div className="px-6 py-4 border-b border-[#eadfd0]">
        <p className="text-[#2b1d16] text-lg font-normal">{title}</p>
        <p className="text-[#7e4722] text-xs mt-0.5">{subtitle}</p>
      </div>
      <div className="bg-[#faf6ee] flex items-center px-6 h-10 gap-4 border-b border-[#eadfd0]">
        <p className="w-[72px] text-xs font-medium text-[#7e4722]">Urutan</p>
        <p className="flex-1 min-w-0 text-xs font-medium text-[#7e4722]">Pesanan / pembeli</p>
        <p className="w-[150px] text-xs font-medium text-[#7e4722]">Jam kirim</p>
        <p className="w-[64px] text-xs font-medium text-[#7e4722]">Qty box</p>
        <p className="w-[80px] text-xs font-medium text-[#7e4722]">Beban</p>
        <p className="flex-1 min-w-0 text-xs font-medium text-[#7e4722]">Waktu masuk</p>
        <p className="w-[64px] text-xs font-medium text-[#7e4722]">Skor</p>
      </div>
      {rows.length === 0 && <div className="px-6 py-8 text-center text-[#8a7768] text-sm">Belum ada pesanan di sesi ini.</div>}
      {rows.map((r, idx) => (
        <div key={r.id} className="flex items-center px-6 h-16 gap-4 border-b border-[#eadfd0] last:border-0 hover:bg-[#faf6ee]/50 transition-colors break-inside-avoid">
          <p className="w-[72px] text-sm text-[#2b1d16]">{idx + 1}</p>
          <div className="flex-1 min-w-0">
            <p className="text-sm font-semibold text-[#2b1d16]">{r.kode}</p>
            <p className="text-xs text-[#8a7768] truncate">
              {r.customer || '—'}
              {penanggungJawab(r).label && <> · PJ: {penanggungJawab(r).label}</>}
            </p>
          </div>
          <p className="w-[150px] text-sm text-[#7e4722]">{fmtTgl(r.deliver_at)} · {r.deliver_at.slice(11, 16)}</p>
          <p className="w-[64px] text-sm text-[#2b1d16]">{r.qty_box}</p>
          <p className="w-[80px] text-sm text-[#2b1d16]">{r.qty}</p>
          <p className="flex-1 min-w-0 text-sm text-[#2b1d16]">{fmtTgl(r.ordered_at)}, {r.ordered_at.slice(11, 16)}</p>
          <div className="w-[64px]"><ScoreBadge score={r.score} /></div>
        </div>
      ))}
    </div>
  );
}

export default function JadwalProduksi() {
  const [tanggal, setTanggal] = useState(ymd());
  const { data, error, loading } = useLoad(() => api.getSchedule(tanggal, 1).then((d) => d[0]), [tanggal]);

  const orders = data?.orders ?? [];
  // Sesi H-1: diproduksi hari ini untuk kirim di hari berikutnya. Sesi Hari H: produksi dan kirim di hari yang sama.
  const h1 = orders.filter((o) => o.deliver_at.slice(0, 10) > tanggal);
  const hariH = orders.filter((o) => o.deliver_at.slice(0, 10) <= tanggal);
  const sum = (rows, k) => rows.reduce((s, r) => s + r[k], 0);

  const capacity = data?.capacity ?? 0;
  const used = data?.used ?? 0;
  const pct = capacity ? Math.round((used / capacity) * 100) : 0;
  const barColor = pct > 90 ? '#b74b40' : pct >= 70 ? '#a66a17' : '#327a52';
  const weekday = new Date(`${tanggal}T00:00:00`).toLocaleDateString('id-ID', { weekday: 'long' });
  const tanggalPanjang = new Date(`${tanggal}T00:00:00`).toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' });

  // Cetak jadwal: pakai dialog print browser (bisa langsung ke printer atau "Simpan sebagai PDF").
  // Judul tab diganti sementara supaya nama file PDF-nya jelas.
  function handleCetak() {
    const judulLama = document.title;
    document.title = `Jadwal Produksi ${tanggal}`;
    window.addEventListener('afterprint', () => { document.title = judulLama; }, { once: true });
    window.print();
  }

  return (
    <div className="flex flex-col min-h-full">
      <div className="flex flex-col gap-6 p-9 flex-1 print:p-0 print:gap-4">
        {/* Kop yang hanya muncul di hasil cetak */}
        <div className="hidden print:block border-b border-[#eadfd0] pb-3">
          <p className="text-[#2b1d16] text-xl font-semibold">Jadwal Produksi Hakuu Bakery</p>
          <p className="text-[#2b1d16] text-sm mt-1">Tanggal produksi: {tanggalPanjang}</p>
          <p className="text-[#8a7768] text-xs mt-0.5">
            Dicetak {new Date().toLocaleString('id-ID', { day: 'numeric', month: 'long', year: 'numeric', hour: '2-digit', minute: '2-digit' })}
          </p>
        </div>

        <div className="print:hidden">
          <h1 className="text-[#2b1d16] text-[30px] font-normal">Jadwal Produksi</h1>
          <p className="text-[#8a7768] text-sm mt-1.5">Susun pekerjaan berdasarkan sesi dan prioritas, bukan urutan pesan masuk saja.</p>
        </div>

        <ErrorBanner message={error} />

        <div className="flex gap-6 items-stretch">
          <div className="card p-5 flex flex-col gap-2 w-[328px] shrink-0 print:hidden">
            <label className="text-[#2b1d16] text-sm">Tanggal produksi</label>
            <input type="date" className="input-field" value={tanggal} onChange={(e) => e.target.value && setTanggal(e.target.value)} />
            <p className="text-[#8a7768] text-xs">{weekday} • jam kerja 06.00–18.00</p>
            <button
              type="button"
              onClick={handleCetak}
              disabled={loading || !data}
              className="btn-secondary justify-center w-full mt-auto text-[#4c0c13] disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <IconPrinter />
              Cetak jadwal
            </button>
          </div>

          <div className="card p-6 flex flex-col gap-3 flex-1">
            <p className="text-[#2b1d16] text-sm">Total kapasitas produksi {tanggal === ymd() ? 'hari ini' : fmtTgl(tanggal)}</p>
            <div className="flex items-center justify-between">
              <span className="text-[#2b1d16] text-lg font-normal">{used} / {capacity} box</span>
              <span className="text-sm font-semibold" style={{ color: barColor }}>{pct}%</span>
            </div>
            <div className="bg-[#eadfd0] rounded-full h-2 overflow-hidden">
              <div className="h-full rounded-full" style={{ width: `${Math.min(pct, 100)}%`, backgroundColor: barColor }} />
            </div>
            <p className="text-[#8a7768] text-xs">
              {sum(h1, 'qty')} box H-1 + {sum(hariH, 'qty')} box hari H • sisa harian {Math.max(0, capacity - used)} box
              {data?.locked ? ' • jadwal terkunci' : ''}
            </p>
          </div>
        </div>

        {loading && !data && <p className="text-[#8a7768] text-sm">Memuat jadwal…</p>}

        <div className="flex flex-col gap-4">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-[#2b1d16] text-lg font-semibold">Sesi H-1</h2>
              <p className="text-[#8a7768] text-xs mt-0.5">
                Produksi {fmtTgl(tanggal)} untuk kirim hari berikutnya • {h1.length} pesanan, {sum(h1, 'qty_box')} box produk
              </p>
            </div>
            <span className="bg-[#eaf5ed] text-[#327a52] text-xs font-medium px-3 py-1.5 rounded-full">{sum(h1, 'qty')} box kapasitas</span>
          </div>
          <SesiTable title="Prioritas sesi H-1" subtitle="Urutan otomatis • skor tertinggi lebih dahulu" rows={h1} />
        </div>

        <div className="flex flex-col gap-4">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-[#2b1d16] text-lg font-semibold">Sesi Hari H</h2>
              <p className="text-[#8a7768] text-xs mt-0.5">
                Produksi dan kirim {fmtTgl(tanggal)} • {hariH.length} pesanan, {sum(hariH, 'qty_box')} box produk
              </p>
            </div>
            <span className="bg-[#eaf5ed] text-[#327a52] text-xs font-medium px-3 py-1.5 rounded-full">{sum(hariH, 'qty')} box kapasitas</span>
          </div>
          <SesiTable title="Prioritas sesi Hari H" subtitle="Urutan otomatis • skor tertinggi lebih dahulu" rows={hariH} />
        </div>
      </div>

      <footer className="border-t border-[#eadfd0] bg-white px-9 py-3 flex items-center justify-between print:hidden">
        <p className="text-[#8a7768] text-xs">Hakuu Bakery • Operasional usaha, bukan etalase pelanggan</p>
        <p className="text-[#8a7768] text-xs">Waktu Indonesia Barat</p>
      </footer>
    </div>
  );
}
