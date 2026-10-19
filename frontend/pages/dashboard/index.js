import { useState } from 'react';
import { useRouter } from 'next/router';
import { iconChevronLeft, iconChevronRight, iconCircleAlert, iconArrowUpRight, iconEllipsis } from '../../utils/icons';
import api from '../../services/bakery';
import { toUiOrder } from '../../utils/mapper';
import { useLoad } from '../../utils/useLoad';
import ErrorBanner from '../../components/ErrorBanner';
import { badgeMap, fmtTgl, isoDate, ymd } from '../../utils/orderStatus';

// Warna bar mengikuti legenda: <70% hijau, 70–90% kuning, >90% merah.
const barColorFor = (pct) => (pct > 90 ? '#b74b40' : pct >= 70 ? '#a66a17' : '#327a52');

function CapacityCard({ day, today }) {
  const pct = day.capacity ? Math.round((day.used / day.capacity) * 100) : 0;
  const isToday = day.date === today;
  const danger = pct > 90;
  const label = new Date(`${day.date}T00:00:00`).toLocaleDateString('id-ID', { weekday: 'short', day: 'numeric', month: 'short' });
  return (
    <div className={`flex flex-col gap-3 p-4 rounded-xl border ${isToday ? 'border-[#4c0c13]' : 'border-[#eadfd0]'} bg-white`}>
      <p className={`text-xs font-medium ${isToday ? 'text-[#4c0c13]' : 'text-[#8a7768]'}`}>{label}</p>
      <div className="flex items-center justify-between">
        <span className={`text-base font-semibold ${danger ? 'text-[#b74b40]' : 'text-[#2b1d16]'}`}>{day.used}</span>
        <span className={`text-xs font-semibold ${danger ? 'text-[#b74b40]' : 'text-[#8a7768]'}`}>{pct}%</span>
      </div>
      <div className="bg-[#eadfd0] rounded-full h-2 overflow-hidden">
        <div className="h-full rounded-full transition-all" style={{ width: `${Math.min(pct, 100)}%`, backgroundColor: barColorFor(pct) }} />
      </div>
    </div>
  );
}

function greeting() {
  const h = new Date().getHours();
  return h < 11 ? 'Selamat pagi' : h < 15 ? 'Selamat siang' : h < 19 ? 'Selamat sore' : 'Selamat malam';
}

export default function Dashboard() {
  const router = useRouter();
  const today = ymd();
  const [weekOffset, setWeekOffset] = useState(0);
  const from = isoDate(weekOffset * 7);
  const to = isoDate(weekOffset * 7 + 6);

  const week = useLoad(() => api.getSchedule(from, 7), [from]);
  const todayLoad = useLoad(() => api.getSchedule(today, 1).then((d) => d[0]), [today]);
  const orders = useLoad(() => api.getOrders(200).then((l) => l.map((o) => toUiOrder(o))), []);

  const list = orders.data || [];
  const live = list.filter((o) => o.statusOtomatis !== 'Dibatalkan');
  const recentOrders = [...live].sort((a, b) => b.waktuMasuk.localeCompare(a.waktuMasuk)).slice(0, 6);

  const prodToday = live.filter((o) => o.tanggalProduksi === today);
  const attention = live.filter((o) => o.needsNego || ['Berisiko', 'Terlambat'].includes(o.statusOtomatis));

  const t = todayLoad.data;
  const used = t?.used ?? 0;
  const capacity = t?.capacity ?? 0;
  const pct = capacity ? Math.round((used / capacity) * 100) : 0;

  const range = `${fmtTgl(from)} – ${fmtTgl(to)}`;
  const attentionNote = (o) => (o.needsNego ? 'menunggu konfirmasi jam kirim' : o.statusOtomatis === 'Berisiko' ? 'belum mulai diproduksi' : 'lewat jam kirim');

  return (
    <div className="flex flex-col min-h-full">
      <div className="flex flex-col gap-6 p-9 flex-1">
        <div className="flex items-start justify-between">
          <div>
            <h1 className="text-[#2b1d16] text-[30px] font-normal leading-tight">{greeting()}, Hakuu Bakery</h1>
            <p className="text-[#8a7768] text-sm mt-1.5">Pantau kapasitas dan pesanan agar produksi hari ini berjalan lancar.</p>
          </div>
          <button onClick={() => router.push('/input-pesanan')} className="btn-primary">+ Input Pesanan</button>
        </div>

        <ErrorBanner message={week.error || todayLoad.error || orders.error} />

        <div className="card p-6">
          <div className="flex items-center justify-between mb-4">
            <p className="text-[#2b1d16] text-sm font-semibold">Kapasitas produksi 7 hari</p>
            <div className="flex gap-4 items-center">
              <button onClick={() => setWeekOffset((w) => w - 1)} className="hover:opacity-60 transition-opacity">
                <img src={iconChevronLeft} alt="Sebelumnya" className="size-4" />
              </button>
              <span className="text-[#2b1d16] text-sm font-medium">{range}</span>
              <button onClick={() => setWeekOffset((w) => w + 1)} className="hover:opacity-60 transition-opacity">
                <img src={iconChevronRight} alt="Berikutnya" className="size-4" />
              </button>
            </div>
          </div>

          <div className="grid grid-cols-7 gap-3">
            {(week.data || []).map((d) => <CapacityCard key={d.date} day={d} today={today} />)}
          </div>

          <div className="flex gap-5 mt-4 items-center">
            {[['#327a52', '<70% · Tersedia'], ['#a66a17', '70–90% · Mendekati penuh'], ['#b74b40', '>90% · Perlu perhatian']].map(([c, l]) => (
              <div key={l} className="flex items-center gap-1.5">
                <div className="size-2 rounded-full" style={{ backgroundColor: c }} />
                <span className="text-[#8a7768] text-xs">{l}</span>
              </div>
            ))}
            <span className="text-[#8a7768] text-xs ml-2">Beban tertimbang dalam box kapasitas</span>
          </div>
        </div>

        <div className="grid grid-cols-4 gap-4">
          <div className="card p-6 flex flex-col gap-2">
            <p className="text-[#8a7768] text-xs">Kapasitas hari ini</p>
            <p className="text-[#2b1d16] text-[30px] font-normal leading-tight">{used} / {capacity}</p>
            <p className="text-[#8a7768] text-xs">box kapasitas · {pct}% terpakai</p>
            <div className="bg-[#eadfd0] rounded-full h-2 mt-1 overflow-hidden">
              <div className="h-full rounded-full" style={{ width: `${Math.min(pct, 100)}%`, backgroundColor: barColorFor(pct) }} />
            </div>
          </div>
          <div className="card p-6 flex flex-col gap-2">
            <p className="text-[#8a7768] text-xs">Sisa kapasitas</p>
            <p className="text-[#2b1d16] text-[30px] font-normal leading-tight">{Math.max(0, capacity - used)} box</p>
            <p className="text-[#8a7768] text-xs">Ruang produksi yang masih tersedia</p>
          </div>
          <div className="card p-6 flex flex-col gap-2">
            <p className="text-[#8a7768] text-xs">Total pesanan</p>
            <p className="text-[#2b1d16] text-[30px] font-normal leading-tight">{prodToday.length} pesanan</p>
            <p className="text-[#8a7768] text-xs">{prodToday.reduce((s, o) => s + o.qty, 0)} box produk · produksi hari ini</p>
          </div>
          <div className="card p-6 flex flex-col gap-2">
            <p className="text-[#8a7768] text-xs">Pesanan perlu perhatian</p>
            <p className="text-[#2b1d16] text-[30px] font-normal leading-tight">{attention.length} pesanan</p>
            <p className="text-[#8a7768] text-xs">Konfirmasi jam kirim, belum mulai, atau terlambat</p>
          </div>
        </div>

        {attention.length > 0 && (
          <div className="bg-[#fdf6ec] border border-[#e8c98a] rounded-xl px-5 py-3.5 flex items-center justify-between">
            <div className="flex items-center gap-3">
              <img src={iconCircleAlert} alt="" className="size-5 shrink-0" style={{ filter: 'hue-rotate(20deg) saturate(1.5)' }} />
              <p className="text-[#7e4722] text-sm">
                {attention.slice(0, 3).map((o) => `${o.id} ${attentionNote(o)}`).join(' · ')}
                {attention.length > 3 ? ` · +${attention.length - 3} lainnya` : ''}
              </p>
            </div>
            <button onClick={() => router.push('/daftar-pesanan')} className="text-[#7e4722] text-sm font-semibold whitespace-nowrap hover:underline">
              Periksa pesanan →
            </button>
          </div>
        )}

        <div className="card">
          <div className="flex items-center justify-between px-6 py-5 border-b border-[#eadfd0]">
            <div>
              <p className="text-[#2b1d16] text-lg font-semibold">Pesanan terbaru</p>
              <p className="text-[#7e4722] text-xs mt-0.5">{recentOrders.length} pesanan terakhir masuk</p>
            </div>
            <button onClick={() => router.push('/daftar-pesanan')} className="flex gap-2 items-center border border-[#eadfd0] rounded-lg px-3 h-9 text-sm text-[#2b1d16] hover:bg-[#faf6ee] transition-colors">
              <img src={iconArrowUpRight} alt="" className="size-4" />
              <span>Lihat semua pesanan</span>
            </button>
          </div>

          <table className="w-full">
            <thead>
              <tr className="bg-[#faf6ee] border-b border-[#eadfd0]">
                {['Kode', 'Pembeli / sumber', 'Produk', 'Qty box', 'Kirim / produksi', 'Status', 'Aksi'].map((h) => (
                  <th key={h} className="px-6 py-3 text-left text-xs font-medium text-[#7e4722]">{h}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {recentOrders.length === 0 && (
                <tr><td colSpan={7} className="px-6 py-8 text-center text-[#8a7768] text-sm">{orders.loading ? 'Memuat…' : 'Belum ada pesanan.'}</td></tr>
              )}
              {recentOrders.map((o) => (
                <tr key={o.id} className="border-b border-[#eadfd0] last:border-0 hover:bg-[#faf6ee]/60 transition-colors">
                  <td className="px-6 py-4 text-sm font-semibold text-[#2b1d16]">{o.id}</td>
                  <td className="px-6 py-4">
                    <p className="text-sm text-[#2b1d16]">{o.pembeli}</p>
                    {o.pj && <p className="text-xs text-[#8a7768]" title={o.pjEmail || o.pj}>PJ: {o.pj}</p>}
                  </td>
                  <td className="px-6 py-4 text-sm font-medium text-[#4c0c13]">{o.produk}</td>
                  <td className="px-6 py-4 text-sm text-[#2b1d16]">{o.qty}</td>
                  <td className="px-6 py-4">
                    <p className="text-sm text-[#2b1d16]">{fmtTgl(o.tanggalKirim)} · {o.jamKirim}</p>
                    <p className="text-xs text-[#8a7768]">{o.tanggalProduksi ? fmtTgl(o.tanggalProduksi) : 'Menunggu negosiasi'}</p>
                  </td>
                  <td className="px-6 py-4"><span className={badgeMap[o.statusOtomatis] ?? badgeMap['Diterima']}>{o.statusOtomatis}</span></td>
                  <td className="px-6 py-4">
                    <button onClick={() => router.push('/daftar-pesanan')} className="text-[#8a7768] hover:text-[#2b1d16]">
                      <img src={iconEllipsis} alt="Buka daftar pesanan" className="size-4" />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      <footer className="border-t border-[#eadfd0] bg-white px-9 py-3 flex items-center justify-between">
        <p className="text-[#8a7768] text-xs">Hakuu Bakery · Operasional usaha, bukan etalase pelanggan</p>
        <p className="text-[#8a7768] text-xs">Waktu Indonesia Barat</p>
      </footer>
    </div>
  );
}
