import { useMemo, useState } from 'react';
import api from '../../services/bakery';
import { useLoad } from '../../utils/useLoad';
import { toUiOrder, penanggungJawab } from '../../utils/mapper';
import { badgeMap, fmtTgl, statusText, ymd } from '../../utils/orderStatus';
import ErrorBanner from '../../components/ErrorBanner';

/*
   note : Laporan Penjualan bulanan
          - daftar pesanan diambil dari /api/orders, lalu disaring per bulan (berdasarkan tanggal kirim)
          - alamat, rincian produk & total harga diambil dari detail /api/orders/:id
          - pesanan "Dibatalkan" tidak dihitung ke produksi maupun penghasilan
          - ringkasan resmi, kinerja penjadwalan, unduh CSV & tutup bulan dari /api/reports/monthly
            (backend hanya menghitung pesanan berstatus selesai)
*/

const rp = (n) => `Rp ${Math.round(Number(n) || 0).toLocaleString('id-ID')}`;
const num = (n) => (Number(n) || 0).toLocaleString('id-ID');

/** 'YYYY-MM' bulan ini */
const thisMonth = () => ymd().slice(0, 7);

/** geser bulan 'YYYY-MM' sebanyak delta */
function shiftMonth(ym, delta) {
  const [y, m] = ym.split('-').map(Number);
  const d = new Date(y, m - 1 + delta, 1);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`;
}

const pctLabel = (v) => (v == null ? '—' : `${Number(v).toLocaleString('id-ID', { maximumFractionDigits: 1 })}%`);

const monthLabel = (ym) => {
  const [y, m] = ym.split('-').map(Number);
  return new Date(y, m - 1, 1).toLocaleDateString('id-ID', { month: 'long', year: 'numeric' });
};

/** jalankan fn untuk tiap item, maksimal `limit` request bersamaan */
async function mapLimit(list, limit, fn, onProgress) {
  const out = new Array(list.length);
  let next = 0;
  let done = 0;
  async function worker() {
    while (next < list.length) {
      const i = next++;
      out[i] = await fn(list[i]);
      done += 1;
      onProgress?.(done, list.length);
    }
  }
  await Promise.all(Array.from({ length: Math.min(limit, list.length) }, worker));
  return out;
}

function IconPrinter() {
  return (
    <svg viewBox="0 0 24 24" className="size-4 shrink-0" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M6 9V2h12v7" />
      <path d="M6 18H4a2 2 0 0 1-2-2v-5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v5a2 2 0 0 1-2 2h-2" />
      <rect x="6" y="14" width="12" height="8" />
    </svg>
  );
}

/** Ringkasan resmi dari server: pesanan selesai, perbandingan bulan lalu, kinerja penjadwalan */
function RingkasanServer({ report, loading, error }) {
  if (error) return <ErrorBanner message={`Ringkasan server gagal dimuat: ${error}`} />;
  if (!report) return loading ? <p className="text-[#8a7768] text-sm">Memuat ringkasan server…</p> : null;

  const s = report.summary;
  const sc = report.scheduling;
  const chg = report.revenue_change_pct;
  const status = report.order_status || {};
  const metrik = [
    { label: 'Tepat waktu', value: pctLabel(sc.on_time_pct), note: `${num(sc.on_time_orders)} dari ${num(sc.delivered_orders)} diserahkan` },
    { label: 'Pemakaian kapasitas', value: pctLabel(sc.capacity_utilization_pct), note: `Lembur ${num(sc.overtime_days)} hari · ${num(sc.overtime_pcs)} unit` },
    { label: 'Pesanan digeser / dimajukan', value: `${num(sc.orders_moved)} / ${num(sc.orders_advanced)}`, note: 'Penyesuaian jadwal otomatis' },
    { label: 'Negosiasi jam kirim', value: `${num(sc.negotiation_accepted)} / ${num(sc.negotiation_offered)} diterima`, note: `${num(sc.negotiation_declined_or_expired)} ditolak / kedaluwarsa` },
    { label: 'Pesanan ditolak', value: num(sc.orders_rejected), note: 'Kapasitas tidak tersedia' },
  ];

  return (
    <div className="card break-inside-avoid">
      <div className="flex items-start justify-between gap-4 px-6 py-4 border-b border-[#eadfd0]">
        <div>
          <p className="text-[#2b1d16] text-lg font-normal">Ringkasan resmi &amp; kinerja penjadwalan</p>
          <p className="text-[#7e4722] text-xs mt-0.5">Dihitung server dari pesanan berstatus selesai</p>
        </div>
        {report.closed
          ? <span className="bg-[#dcfce7] text-[#166534] text-xs font-medium px-2.5 py-1 rounded-lg shrink-0">Ditutup {report.closed_at}</span>
          : <span className="bg-[#fcebc8] text-[#7e4722] text-xs font-medium px-2.5 py-1 rounded-lg shrink-0">Belum ditutup</span>}
      </div>
      <div className="grid grid-cols-4 gap-4 px-6 py-5 border-b border-[#eadfd0]">
        <div>
          <p className="text-[#8a7768] text-xs">Penghasilan (selesai)</p>
          <p className="text-[#2b1d16] text-xl">{rp(s.revenue)}</p>
          <p className={`text-xs ${chg == null ? 'text-[#8a7768]' : chg >= 0 ? 'text-[#327a52]' : 'text-[#b74b40]'}`}>
            {chg == null ? 'Bulan lalu belum ada penjualan' : `${chg >= 0 ? '▲' : '▼'} ${pctLabel(Math.abs(chg))} dari bulan lalu (${rp(report.previous_month.revenue)})`}
          </p>
        </div>
        <div>
          <p className="text-[#8a7768] text-xs">Pesanan selesai</p>
          <p className="text-[#2b1d16] text-xl">{num(s.orders)}</p>
          <p className="text-[#8a7768] text-xs">{num(s.pcs)} unit kapasitas</p>
        </div>
        <div>
          <p className="text-[#8a7768] text-xs">Rata-rata per pesanan</p>
          <p className="text-[#2b1d16] text-xl">{rp(s.avg_order_value)}</p>
        </div>
        <div>
          <p className="text-[#8a7768] text-xs">Status pesanan bulan ini</p>
          <p className="text-[#2b1d16] text-sm mt-1">
            {Object.keys(status).length
              ? Object.entries(status).map(([k, v]) => `${k.replaceAll('_', ' ')} ${v}`).join(' · ')
              : '—'}
          </p>
        </div>
      </div>
      <div className="grid grid-cols-5 gap-4 px-6 py-5">
        {metrik.map((m) => (
          <div key={m.label}>
            <p className="text-[#8a7768] text-xs">{m.label}</p>
            <p className="text-[#2b1d16] text-base font-semibold mt-0.5">{m.value}</p>
            <p className="text-[#8a7768] text-xs mt-0.5">{m.note}</p>
          </div>
        ))}
      </div>
    </div>
  );
}

function StatCard({ label, value, note, accent = false }) {
  return (
    <div className={`card p-5 flex flex-col gap-1.5 flex-1 min-w-[200px] ${accent ? 'bg-[#4c0c13] border-[#4c0c13]' : ''}`}>
      <p className={`text-xs ${accent ? 'text-[#fcebc8]' : 'text-[#8a7768]'}`}>{label}</p>
      <p className={`text-[26px] leading-tight font-normal ${accent ? 'text-white' : 'text-[#2b1d16]'}`}>{value}</p>
      {note && <p className={`text-xs ${accent ? 'text-[#fcebc8]/80' : 'text-[#7e4722]'}`}>{note}</p>}
    </div>
  );
}

export default function LaporanPenjualan() {
  const [bulan, setBulan] = useState(thisMonth());
  const [search, setSearch] = useState('');
  const [showBatal, setShowBatal] = useState(false);
  const [progress, setProgress] = useState(null);

  const { data, error, loading, reload } = useLoad(async () => {
    setProgress(null);
    const list = await api.getOrders(1000);
    const bulanIni = (list || []).filter((o) => (o.deliver_at || '').slice(0, 7) === bulan);
    const details = await mapLimit(bulanIni, 6, async (o) => {
      try { return await api.getOrder(o.id); } catch { return null; }
    }, (done, total) => setProgress({ done, total }));
    setProgress(null);
    return bulanIni.map((o, i) => {
      const ui = toUiOrder(o);
      const d = details[i];
      const items = d?.items ?? [];
      const pjDetail = d ? penanggungJawab(d) : { label: '', email: '' };
      return {
        ...ui,
        pj: pjDetail.label || ui.pj,
        pjEmail: pjDetail.email || ui.pjEmail,
        telepon: d?.phone || '',
        alamat: d?.address || '',
        items,
        totalHarga: Number(d?.total_harga ?? items.reduce((s, it) => s + (Number(it.subtotal) || 0), 0)),
        detailGagal: !d,
      };
    }).sort((a, b) => (a.tanggalKirim + a.jamKirim).localeCompare(b.tanggalKirim + b.jamKirim));
  }, [bulan]);

  const report = useLoad(() => api.getMonthlyReport(bulan), [bulan]);
  const [aksiErr, setAksiErr] = useState('');
  const [aksiBusy, setAksiBusy] = useState('');

  async function handleUnduhCsv() {
    setAksiErr('');
    setAksiBusy('csv');
    try {
      const blob = await api.exportMonthlyReport(bulan);
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `laporan-${bulan}.csv`;
      a.click();
      URL.revokeObjectURL(url);
    } catch (e) {
      setAksiErr(e.message);
    } finally {
      setAksiBusy('');
    }
  }

  async function handleTutupBulan() {
    if (!window.confirm(`Tutup laporan ${monthLabel(bulan)}? Ringkasan akan disimpan permanen dan tidak berubah lagi.`)) return;
    setAksiErr('');
    setAksiBusy('tutup');
    try {
      await api.closeMonth(bulan);
      report.reload();
    } catch (e) {
      setAksiErr(e.message);
    } finally {
      setAksiBusy('');
    }
  }

  const rows = data || [];
  const sah = rows.filter((o) => o.statusOtomatis !== 'Dibatalkan');
  const batal = rows.length - sah.length;
  const selesai = sah.filter((o) => o.statusOtomatis === 'Selesai');

  const totalBox = sah.reduce((s, o) => s + o.qty, 0);
  const totalHarga = sah.reduce((s, o) => s + o.totalHarga, 0);
  const totalSelesai = selesai.reduce((s, o) => s + o.totalHarga, 0);
  const rata = sah.length ? totalHarga / sah.length : 0;

  // rekap per produk (dari rincian item tiap pesanan yang tidak dibatalkan)
  const perProduk = useMemo(() => {
    const map = {};
    sah.forEach((o) => o.items.forEach((it) => {
      const k = it.produk || '—';
      map[k] = map[k] || { produk: k, box: 0, nilai: 0, pesanan: 0 };
      map[k].box += Number(it.qty) || 0;
      map[k].nilai += Number(it.subtotal) || 0;
      map[k].pesanan += 1;
    }));
    return Object.values(map).sort((a, b) => b.box - a.box);
  }, [data]);
  const maxBox = perProduk[0]?.box || 1;
  const terlaris = perProduk[0];

  const tampil = useMemo(() => {
    const q = search.trim().toLowerCase();
    return rows
      .filter((o) => showBatal || o.statusOtomatis !== 'Dibatalkan')
      .filter((o) => !q || [o.id, o.pembeli, o.alamat, o.produk, o.pj, o.pjEmail].some((v) => (v || '').toLowerCase().includes(q)));
  }, [data, search, showBatal]);
  const tampilSah = tampil.filter((o) => o.statusOtomatis !== 'Dibatalkan');

  const isCurrent = bulan === thisMonth();
  const bisaTutup = bulan < thisMonth() && report.data && !report.data.closed;
  const adaGagal = rows.some((o) => o.detailGagal);

  function handleCetak() {
    const judulLama = document.title;
    document.title = `Laporan Penjualan ${bulan}`;
    window.addEventListener('afterprint', () => { document.title = judulLama; }, { once: true });
    window.print();
  }

  return (
    <div className="flex flex-col min-h-full">
      <div className="flex flex-col gap-6 p-9 flex-1 print:p-0 print:gap-4">
        {/* Kop khusus hasil cetak */}
        <div className="hidden print:block border-b border-[#eadfd0] pb-3">
          <p className="text-[#2b1d16] text-xl font-semibold">Laporan Penjualan Hakuu Bakery</p>
          <p className="text-[#2b1d16] text-sm mt-1">Periode: {monthLabel(bulan)} (berdasarkan tanggal kirim)</p>
          <p className="text-[#8a7768] text-xs mt-0.5">
            Dicetak {new Date().toLocaleString('id-ID', { day: 'numeric', month: 'long', year: 'numeric', hour: '2-digit', minute: '2-digit' })}
          </p>
        </div>

        {/* Judul + aksi */}
        <div className="flex items-start justify-between gap-6 print:hidden">
          <div>
            <h1 className="text-[#2b1d16] text-[30px] font-normal">Laporan Penjualan</h1>
            <p className="text-[#8a7768] text-sm mt-1.5">Rekap produksi dan penghasilan per bulan, dihitung dari tanggal kirim pesanan.</p>
          </div>
          <div className="flex gap-3 shrink-0">
            {bisaTutup && (
              <button onClick={handleTutupBulan} disabled={!!aksiBusy}
                className="btn-secondary text-[#4c0c13] disabled:opacity-50">
                {aksiBusy === 'tutup' ? 'Menutup…' : 'Tutup bulan'}
              </button>
            )}
            <button onClick={handleUnduhCsv} disabled={!!aksiBusy}
              className="btn-secondary text-[#4c0c13] disabled:opacity-50">
              {aksiBusy === 'csv' ? 'Mengunduh…' : 'Unduh CSV'}
            </button>
            <button onClick={handleCetak} disabled={loading || !data}
              className="btn-primary disabled:opacity-50 disabled:cursor-not-allowed">
              <IconPrinter /> Cetak laporan
            </button>
          </div>
        </div>

        {/* Pilih bulan */}
        <div className="card p-5 flex items-center gap-4 flex-wrap print:hidden">
          <p className="text-[#2b1d16] text-sm">Periode</p>
          <div className="flex items-center gap-2">
            <button onClick={() => setBulan(shiftMonth(bulan, -1))} aria-label="Bulan sebelumnya"
              className="size-11 rounded-xl border border-[#eadfd0] text-[#4c0c13] text-lg hover:bg-[#faf6ee]">‹</button>
            <input type="month" className="input-field w-[190px]" value={bulan}
              onChange={(e) => e.target.value && setBulan(e.target.value)} />
            <button onClick={() => setBulan(shiftMonth(bulan, 1))} aria-label="Bulan berikutnya"
              className="size-11 rounded-xl border border-[#eadfd0] text-[#4c0c13] text-lg hover:bg-[#faf6ee]">›</button>
          </div>
          <p className="text-[#2b1d16] text-lg font-normal ml-2">{monthLabel(bulan)}</p>
          {!isCurrent && (
            <button onClick={() => setBulan(thisMonth())} className="text-[#4c0c13] text-xs font-semibold hover:underline">Kembali ke bulan ini</button>
          )}
          <button onClick={() => { reload(); report.reload(); }} className="ml-auto border border-[#eadfd0] rounded-lg px-3 h-9 text-sm text-[#2b1d16] font-semibold hover:bg-[#faf6ee]">
            Muat ulang
          </button>
        </div>

        <ErrorBanner message={error || aksiErr} />
        {adaGagal && (
          <div className="bg-[#fdf6ec] border border-[#e8c98a] rounded-xl px-5 py-3 text-sm text-[#a66a17] print:hidden">
            Rincian beberapa pesanan gagal dimuat, sehingga alamat dan total harganya belum lengkap. Tekan Muat ulang untuk mencoba lagi.
          </div>
        )}

        {loading && (
          <p className="text-[#8a7768] text-sm">
            Memuat laporan{progress ? ` · rincian pesanan ${progress.done}/${progress.total}` : '…'}
          </p>
        )}

        {/* Ringkasan atas */}
        <div className="flex gap-4 flex-wrap">
          <StatCard accent label={`Total penghasilan ${monthLabel(bulan)}`} value={rp(totalHarga)}
            note={`${rp(totalSelesai)} dari pesanan selesai`} />
          <StatCard label="Total produksi" value={`${num(totalBox)} box`}
            note={`${perProduk.length} jenis produk`} />
          <StatCard label="Jumlah pesanan" value={`${num(sah.length)} pesanan`}
            note={`${selesai.length} selesai · ${batal} dibatalkan`} />
          <StatCard label="Rata-rata per pesanan" value={rp(rata)}
            note={terlaris ? `Terlaris: ${terlaris.produk}` : 'Belum ada penjualan'} />
        </div>

        <RingkasanServer report={report.data} loading={report.loading} error={report.error} />

        {/* Rekap produksi per produk */}
        <div className="card break-inside-avoid">
          <div className="px-6 py-4 border-b border-[#eadfd0]">
            <p className="text-[#2b1d16] text-lg font-normal">Rekap produksi per produk</p>
            <p className="text-[#7e4722] text-xs mt-0.5">Jumlah box yang diproduksi selama {monthLabel(bulan)}, tanpa pesanan dibatalkan</p>
          </div>
          <div className="bg-[#faf6ee] flex items-center px-6 h-10 gap-4 border-b border-[#eadfd0]">
            <p className="w-[40px] text-xs font-medium text-[#7e4722]">No</p>
            <p className="flex-1 min-w-0 text-xs font-medium text-[#7e4722]">Produk</p>
            <p className="w-[220px] text-xs font-medium text-[#7e4722] print:hidden">Porsi produksi</p>
            <p className="w-[90px] text-xs font-medium text-[#7e4722] text-right">Total box</p>
            <p className="w-[90px] text-xs font-medium text-[#7e4722] text-right">Pesanan</p>
            <p className="w-[140px] text-xs font-medium text-[#7e4722] text-right">Penghasilan</p>
          </div>
          {!loading && perProduk.length === 0 && (
            <div className="px-6 py-8 text-center text-[#8a7768] text-sm">Belum ada produk terjual di bulan ini.</div>
          )}
          {perProduk.map((p, i) => (
            <div key={p.produk} className="flex items-center px-6 h-12 gap-4 border-b border-[#eadfd0] last:border-0 break-inside-avoid">
              <p className="w-[40px] text-sm text-[#8a7768]">{i + 1}</p>
              <p className="flex-1 min-w-0 text-sm font-semibold text-[#2b1d16] truncate">{p.produk}</p>
              <div className="w-[220px] print:hidden">
                <div className="bg-[#eadfd0] rounded-full h-2 overflow-hidden">
                  <div className="h-full rounded-full" style={{ width: `${(p.box / maxBox) * 100}%`, backgroundColor: i === 0 ? '#4c0c13' : '#daa278' }} />
                </div>
              </div>
              <p className="w-[90px] text-sm text-[#2b1d16] text-right">{num(p.box)}</p>
              <p className="w-[90px] text-sm text-[#8a7768] text-right">{num(p.pesanan)}</p>
              <p className="w-[140px] text-sm font-medium text-[#4c0c13] text-right">{rp(p.nilai)}</p>
            </div>
          ))}
          {perProduk.length > 0 && (
            <div className="flex items-center px-6 h-12 gap-4 bg-[#faf6ee] rounded-b-xl">
              <p className="w-[40px]" />
              <p className="flex-1 text-sm font-semibold text-[#2b1d16]">Total</p>
              <p className="w-[220px] print:hidden" />
              <p className="w-[90px] text-sm font-semibold text-[#2b1d16] text-right">{num(totalBox)}</p>
              <p className="w-[90px]" />
              <p className="w-[140px] text-sm font-semibold text-[#4c0c13] text-right">{rp(perProduk.reduce((s, p) => s + p.nilai, 0))}</p>
            </div>
          )}
        </div>

        {/* Tabel pesanan */}
        <div className="card">
          <div className="flex items-end justify-between gap-4 px-6 py-5 border-b border-[#eadfd0] flex-wrap">
            <div>
              <p className="text-[#2b1d16] text-lg font-normal">Data pesanan {monthLabel(bulan)}</p>
              <p className="text-[#7e4722] text-xs mt-0.5">{tampil.length} pesanan ditampilkan · urut tanggal kirim</p>
            </div>
            <div className="flex items-center gap-4 print:hidden">
              <label className="flex items-center gap-2 text-sm text-[#2b1d16] cursor-pointer select-none">
                <input type="checkbox" className="size-4 accent-[#4c0c13]" checked={showBatal} onChange={(e) => setShowBatal(e.target.checked)} />
                Tampilkan dibatalkan
              </label>
              <input className="input-field w-[260px]" placeholder="Cari nama, alamat, kode, produk, PJ"
                value={search} onChange={(e) => setSearch(e.target.value)} />
            </div>
          </div>

          <div className="overflow-x-auto print:overflow-visible">
            <table className="w-full text-sm border-collapse">
              <thead>
                <tr className="bg-[#faf6ee] text-left text-xs font-medium text-[#7e4722]">
                  <th className="px-6 py-3 font-medium w-[48px]">No</th>
                  <th className="px-3 py-3 font-medium whitespace-nowrap">Tanggal kirim</th>
                  <th className="px-3 py-3 font-medium">Nama pemesan</th>
                  <th className="px-3 py-3 font-medium min-w-[200px]">Alamat</th>
                  <th className="px-3 py-3 font-medium min-w-[200px]">Produk dipesan</th>
                  <th className="px-3 py-3 font-medium whitespace-nowrap">Penanggung jawab</th>
                  <th className="px-3 py-3 font-medium text-right whitespace-nowrap">Total box</th>
                  <th className="px-3 py-3 font-medium text-right whitespace-nowrap">Total pesanan</th>
                  <th className="px-6 py-3 font-medium">Status</th>
                </tr>
              </thead>
              <tbody>
                {!loading && tampil.length === 0 && (
                  <tr><td colSpan={9} className="px-6 py-10 text-center text-[#8a7768]">
                    {rows.length === 0 ? 'Belum ada pesanan dengan tanggal kirim di bulan ini.' : 'Tidak ada pesanan yang cocok dengan pencarian.'}
                  </td></tr>
                )}
                {tampil.map((o, i) => {
                  const dibatalkan = o.statusOtomatis === 'Dibatalkan';
                  return (
                    <tr key={o.id} className={`border-t border-[#eadfd0] align-top break-inside-avoid hover:bg-[#faf6ee]/50 ${dibatalkan ? 'text-[#8a7768]' : 'text-[#2b1d16]'}`}>
                      <td className="px-6 py-3 text-[#8a7768]">{i + 1}</td>
                      <td className="px-3 py-3 whitespace-nowrap">
                        <p>{fmtTgl(o.tanggalKirim)} · {o.jamKirim}</p>
                        <p className="text-xs text-[#8a7768]">{o.id}</p>
                      </td>
                      <td className="px-3 py-3">
                        <p className="font-semibold">{o.pembeli}</p>
                        {o.telepon && <p className="text-xs text-[#8a7768]">{o.telepon}</p>}
                      </td>
                      <td className="px-3 py-3 text-[#78685c]">{o.alamat || '—'}</td>
                      <td className="px-3 py-3">
                        {o.items.length ? (
                          <ul className="flex flex-col gap-0.5">
                            {o.items.map((it, k) => (
                              <li key={k}><span className={dibatalkan ? '' : 'text-[#4c0c13] font-medium'}>{it.produk}</span> <span className="text-[#8a7768]">×{it.qty}</span></li>
                            ))}
                          </ul>
                        ) : <span>{o.produk}</span>}
                      </td>
                      <td className="px-3 py-3">
                        <p>{o.pj || '—'}</p>
                        {o.pjEmail && o.pjEmail !== o.pj && <p className="text-xs text-[#8a7768]">{o.pjEmail}</p>}
                      </td>
                      <td className="px-3 py-3 text-right whitespace-nowrap">{num(o.qty)}</td>
                      <td className={`px-3 py-3 text-right whitespace-nowrap font-semibold ${dibatalkan ? 'line-through' : 'text-[#4c0c13]'}`}>{rp(o.totalHarga)}</td>
                      <td className="px-6 py-3 whitespace-nowrap">
                        <span className={badgeMap[o.statusOtomatis] ?? badgeMap['Diterima']}>{statusText(o)}</span>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
              {tampilSah.length > 0 && (
                <tfoot>
                  <tr className="border-t border-[#eadfd0] bg-[#faf6ee]">
                    <td colSpan={6} className="px-6 py-3.5 font-semibold text-[#2b1d16]">
                      Total {tampilSah.length} pesanan{showBatal && batal ? ' (tanpa dibatalkan)' : ''}
                    </td>
                    <td className="px-3 py-3.5 text-right font-semibold text-[#2b1d16] whitespace-nowrap">{num(tampilSah.reduce((s, o) => s + o.qty, 0))}</td>
                    <td className="px-3 py-3.5 text-right font-semibold text-[#4c0c13] whitespace-nowrap">{rp(tampilSah.reduce((s, o) => s + o.totalHarga, 0))}</td>
                    <td className="px-6 py-3.5" />
                  </tr>
                </tfoot>
              )}
            </table>
          </div>
        </div>
      </div>

      <footer className="border-t border-[#eadfd0] bg-white px-9 py-3 flex items-center justify-between print:hidden">
        <p className="text-[#8a7768] text-xs">Hakuu Bakery • Penghasilan dihitung dari pesanan yang tidak dibatalkan</p>
        <p className="text-[#8a7768] text-xs">Waktu Indonesia Barat</p>
      </footer>
    </div>
  );
}
