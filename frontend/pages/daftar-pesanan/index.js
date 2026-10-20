import { useState, useMemo, useEffect } from 'react';
import { useRouter } from 'next/router';
import { iconArrowUpRight } from '../../utils/icons';
import api from '../../services/bakery';
import { toUiOrder, penanggungJawab } from '../../utils/mapper';
import { useLoad } from '../../utils/useLoad';
import ErrorBanner from '../../components/ErrorBanner';
import { badgeMap, fmtTgl, hLabel, statusText, ymd } from '../../utils/orderStatus';

const STATUS_TABS = ['Semua', 'Diterima', 'Dijadwalkan', 'Diproduksi', 'Selesai', 'Dibatalkan'];
const rp = (n) => `Rp ${Number(n).toLocaleString('id-ID')}`;

/** Aksi manual yang tersedia menurut status backend. */
function actionsFor(o) {
  const a = [];
  if (['diproduksi', 'terlambat'].includes(o.backendStatus)) a.push({ key: 'complete', label: 'Tandai selesai' });
  if (!['selesai', 'batal'].includes(o.backendStatus)) a.push({ key: 'cancel', label: 'Batalkan', danger: true });
  return a;
}

const OFFER_LABEL = {
  direkomendasikan: 'Rekomendasi',
  ditawarkan: 'Menunggu jawaban pembeli',
  diterima: 'Diterima pembeli',
  ditolak: 'Ditolak pembeli',
  kedaluwarsa: 'Kedaluwarsa',
};

function IconPrinter() {
  return (
    <svg viewBox="0 0 24 24" className="size-4 shrink-0" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M6 9V2h12v7" />
      <path d="M6 18H4a2 2 0 0 1-2-2v-5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v5a2 2 0 0 1-2 2h-2" />
      <rect x="6" y="14" width="12" height="8" />
    </svg>
  );
}

/**
 * Rekap produksi satu hari: total box per produk dari semua pesanan yang diproduksi hari itu.
 * Paket pilih varian dirinci per varian (tiap box berisi satu dari setiap varian yang dipilih).
 */
function rekapProduk(details) {
  const map = new Map();
  details.forEach((d) => d.items.forEach((it) => {
    const r = map.get(it.produk) || { produk: it.produk, box: 0, varian: new Map() };
    r.box += it.qty;
    (it.variants || []).forEach((v) => {
      const k = v.opsi ? `${v.nama} (${v.opsi})` : v.nama;
      r.varian.set(k, (r.varian.get(k) || 0) + it.qty);
    });
    map.set(it.produk, r);
  }));
  return [...map.values()].sort((a, b) => a.produk.localeCompare(b.produk));
}

/** Lembar rekap yang hanya muncul di hasil cetak. */
function RekapCetak({ tanggal, details }) {
  const rows = rekapProduk(details);
  const totalBox = rows.reduce((s, r) => s + r.box, 0);
  const tanggalPanjang = new Date(`${tanggal}T00:00:00`).toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' });
  return (
    <div className="hidden print:flex flex-col gap-5 p-2 text-sm">
      <div className="border-b border-[#eadfd0] pb-3">
        <p className="text-[#2b1d16] text-xl font-semibold">Rekap Produksi Hakuu Bakery</p>
        <p className="text-[#2b1d16] text-sm mt-1">Tanggal produksi: {tanggalPanjang} · {details.length} pesanan</p>
        <p className="text-[#8a7768] text-xs mt-0.5">
          Dicetak {new Date().toLocaleString('id-ID', { day: 'numeric', month: 'long', year: 'numeric', hour: '2-digit', minute: '2-digit' })}
        </p>
      </div>
      <table className="w-full border border-[#eadfd0]">
        <thead>
          <tr className="bg-[#faf6ee] text-xs text-[#7e4722] text-left">
            <th className="px-3 py-2 font-medium w-10">No</th>
            <th className="px-3 py-2 font-medium">Produk</th>
            <th className="px-3 py-2 font-medium w-28 text-right">Jumlah box</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((r, i) => (
            <tr key={r.produk} className="border-t border-[#eadfd0] align-top">
              <td className="px-3 py-2 text-[#2b1d16]">{i + 1}</td>
              <td className="px-3 py-2">
                <p className="text-[#2b1d16]">{r.produk}</p>
                {r.varian.size > 0 && (
                  <p className="text-xs text-[#7e4722]">{[...r.varian].map(([k, n]) => `${k} ${n}`).join(' · ')}</p>
                )}
              </td>
              <td className="px-3 py-2 text-right text-[#2b1d16]">{r.box}</td>
            </tr>
          ))}
          <tr className="border-t border-[#eadfd0] font-semibold">
            <td />
            <td className="px-3 py-2 text-[#2b1d16]">Total</td>
            <td className="px-3 py-2 text-right text-[#4c0c13]">{totalBox}</td>
          </tr>
        </tbody>
      </table>
    </div>
  );
}

function DetailModal({ order, onClose, onChanged }) {
  const { data: d, error, loading, reload } = useLoad(() => api.getOrder(order.dbId), [order.dbId]);
  const pj = d ? penanggungJawab(d) : null;
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState('');

  // DEBUG sementara
  useEffect(() => {
    console.log('DETAIL', order.dbId, order.id, d);
    if (d) console.log('ITEMS', d.items, 'isArray:', Array.isArray(d.items), 'length:', d.items?.length);
  }, [d]);

  async function offerAction(fn) {
    setBusy(true);
    setErr('');
    try {
      await fn();
      await reload();
      onChanged();
    } catch (e) {
      setErr(e.message);
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
      <div className="bg-white rounded-2xl shadow-modal w-[600px] flex flex-col max-h-[90vh]">
        <div className="flex items-center justify-between px-6 py-5 border-b border-[#eadfd0]">
          <div>
            <h2 className="text-[#2b1d16] text-lg font-semibold">Detail pesanan {order.id}</h2>
            <span className={`${badgeMap[order.statusOtomatis]} mt-1.5`}>{statusText(order)}</span>
          </div>
          <button onClick={onClose} className="text-[#8a7768] hover:text-[#2b1d16] text-xl leading-none">×</button>
        </div>

        <div className="flex flex-col gap-5 p-6 overflow-y-auto [&>*]:shrink-0">
          {loading && <p className="text-[#8a7768] text-sm">Memuat…</p>}
          <ErrorBanner message={error || err} />
          {d && (
            <>
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div><p className="text-[#8a7768] text-xs">Nama pembeli</p><p className="text-[#2b1d16]">{d.customer || '—'}</p></div>
                <div><p className="text-[#8a7768] text-xs">Nomor telepon</p><p className="text-[#2b1d16]">{d.phone || '—'}</p></div>
                <div className="col-span-2"><p className="text-[#8a7768] text-xs">Alamat</p><p className="text-[#2b1d16]">{d.address || '—'}</p></div>
                <div><p className="text-[#8a7768] text-xs">Kirim</p><p className="text-[#2b1d16]">{fmtTgl(d.deliver_at)} · {d.deliver_at.slice(11, 16)}</p></div>
                <div><p className="text-[#8a7768] text-xs">Produksi</p><p className="text-[#2b1d16]">{fmtTgl(d.production_day)} · {hLabel(d.deliver_at.slice(0, 10), d.production_day)}</p></div>
                <div className="col-span-2">
                  <p className="text-[#8a7768] text-xs">Penanggung jawab</p>
                  <p className="text-[#2b1d16]">
                    {pj.label || order.pj || 'Belum tercatat'}
                    {pj.nama && pj.email && <span className="text-[#8a7768]"> · {pj.email}</span>}
                  </p>
                </div>
                {d.note && <div className="col-span-2"><p className="text-[#8a7768] text-xs">Catatan</p><p className="text-[#2b1d16]">{d.note}</p></div>}
              </div>

              <div className="border border-[#eadfd0] rounded-xl overflow-hidden shrink-0">
                <div className="bg-[#faf6ee] flex px-4 h-9 items-center text-xs font-medium text-[#7e4722]">
                  <p className="flex-1">Produk</p><p className="w-16">Box</p><p className="w-24 text-right">Subtotal</p>
                </div>
                {d.items.map((it, i) => (
                  <div key={i} className="flex px-4 min-h-11 py-2.5 items-center text-sm border-t border-[#eadfd0]">
                    <div className="flex-1">
                      <p className="text-[#2b1d16]">{it.produk}</p>
                      {it.variants?.length > 0 && (
                        <p className="text-xs text-[#7e4722]">
                          {it.variants.map((v) => (v.opsi ? `${v.nama} (${v.opsi})` : v.nama)).join(', ')}
                        </p>
                      )}
                    </div>
                    <p className="w-16 text-[#2b1d16]">{it.qty}</p>
                    <p className="w-24 text-right text-[#2b1d16]">{rp(it.subtotal)}</p>
                  </div>
                ))}
                <div className="flex px-4 h-11 items-center text-sm border-t border-[#eadfd0] bg-[#faf6ee]">
                  <p className="flex-1 font-semibold text-[#2b1d16]">Total · {d.qty_box} box ({d.qty} box kapasitas)</p>
                  <p className="text-right font-semibold text-[#4c0c13]">{rp(d.total_harga)}</p>
                </div>
              </div>

              {d.offers.length > 0 && (
                <div className="flex flex-col gap-2">
                  <p className="text-[#2b1d16] text-sm font-semibold">Alternatif jam kirim (negosiasi)</p>
                  {d.offers.map((of) => (
                    <div key={of.id} className="border border-[#eadfd0] rounded-xl px-4 py-3 flex items-center gap-3">
                      <div className="flex-1">
                        <p className="text-sm text-[#2b1d16]">{fmtTgl(of.deliver_at)} · {of.deliver_at.slice(11, 16)}</p>
                        <p className="text-xs text-[#8a7768]">
                          {OFFER_LABEL[of.status] || of.status} · produksi {fmtTgl(of.production_day)}
                          {of.respond_before && of.status === 'ditawarkan' && ` · jawab sebelum ${new Date(of.respond_before).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })}`}
                        </p>
                      </div>
                      {of.status === 'direkomendasikan' && (
                        <button disabled={busy} onClick={() => offerAction(() => api.offerToBuyer(of.id))} className="btn-secondary text-[#4c0c13] text-xs h-9 px-3">
                          Tawarkan ke pembeli
                        </button>
                      )}
                      {of.status === 'ditawarkan' && (
                        <>
                          <button disabled={busy} onClick={() => offerAction(() => api.respondOffer(of.id, true))} className="btn-primary text-xs h-9 px-3">Pembeli setuju</button>
                          <button disabled={busy} onClick={() => offerAction(() => api.respondOffer(of.id, false))} className="btn-secondary text-xs h-9 px-3">Menolak</button>
                        </>
                      )}
                    </div>
                  ))}
                </div>
              )}

              {d.history.length > 0 && (
                <div className="flex flex-col gap-2">
                  <p className="text-[#2b1d16] text-sm font-semibold">Riwayat</p>
                  {d.history.map((h, i) => (
                    <p key={i} className="text-xs text-[#8a7768]">
                      {h.waktu} · {h.aksi.replaceAll('_', ' ')}{h.alasan ? ` — ${h.alasan}` : ''}
                    </p>
                  ))}
                </div>
              )}
            </>
          )}
        </div>
        <div className="px-6 py-4 border-t border-[#eadfd0]">
          <button onClick={onClose} className="btn-secondary w-full justify-center">Tutup</button>
        </div>
      </div>
    </div>
  );
}

export default function DaftarPesanan() {
  const router = useRouter();
  const { data: orders, error, loading, reload } = useLoad(
    () => api.getOrders(200).then((l) => l.map((o) => toUiOrder(o))), []);
  const [activeTab, setTab] = useState('Semua');
  const [search, setSearch] = useState('');
  const [tanggal, setTanggal] = useState('');
  const [menuOpen, setMenuOpen] = useState(null);
  const [detail, setDetail] = useState(null);
  const [actionErr, setActionErr] = useState('');
  const [cetak, setCetak] = useState(null);       // { tanggal, details } yang sedang dicetak
  const [mencetak, setMencetak] = useState(false);
  const [now, setNow] = useState(new Date());

  useEffect(() => { const t = setInterval(() => setNow(new Date()), 60000); return () => clearInterval(t); }, []);

  const list = orders || [];

  const inTab = (o, tab) => {
    if (tab === 'Semua') return o.statusOtomatis !== 'Dibatalkan';
    return o.statusOtomatis === tab;
  };

  const tabCounts = useMemo(() => {
    const c = {};
    STATUS_TABS.forEach((k) => { c[k] = list.filter((o) => inTab(o, k)).length; });
    return c;
  }, [list]);

  const filtered = useMemo(() => {
    const q = search.toLowerCase();
    return list
      .filter((o) => inTab(o, activeTab)
        && (!tanggal || o.tanggalKirim === tanggal)
        && (o.id.toLowerCase().includes(q) || o.pembeli.toLowerCase().includes(q)))
      .sort((a, b) => {
        const d = (a.tanggalProduksi || a.tanggalKirim).localeCompare(b.tanggalProduksi || b.tanggalKirim);
        return d !== 0 ? d : a.jamKirim.localeCompare(b.jamKirim);
      });
  }, [list, activeTab, search, tanggal]);

  // Pesanan yang diproduksi hari ini (belum selesai / dibatalkan / terlambat).
  const hariIni = ymd(now);
  const produksiHariIni = list.filter((o) => o.tanggalProduksi === hariIni
    && ['terjadwal', 'at_risk', 'diproduksi'].includes(o.backendStatus));

  // Cetak pesanan hari ini: rekap produk & qty yang harus diproduksi hari ini.
  // Pesanan yang belum mulai otomatis masuk status Diproduksi (aksi "start" di server).
  async function cetakPesanan() {
    setActionErr('');
    setMencetak(true);
    try {
      const gagal = [];
      await Promise.all(produksiHariIni.filter((o) => o.canStart).map((o) =>
        api.orderAction(o.dbId, 'start').catch((e) => gagal.push(`${o.id}: ${e.message}`))));
      if (gagal.length) setActionErr(`Gagal mengubah status: ${gagal.join('; ')}`);
      const details = await Promise.all(produksiHariIni.map((o) => api.getOrder(o.dbId)));
      setCetak({ tanggal: hariIni, details });
    } catch (e) {
      setActionErr(e.message);
    } finally {
      setMencetak(false);
      reload();
    }
  }

  useEffect(() => {
    if (!cetak) return;
    const judulLama = document.title;
    document.title = `Rekap Produksi ${cetak.tanggal}`;
    const selesai = () => { document.title = judulLama; setCetak(null); };
    window.addEventListener('afterprint', selesai, { once: true });
    window.print();
    return () => window.removeEventListener('afterprint', selesai);
  }, [cetak]);

  async function runAction(o, a) {
    setMenuOpen(null);
    setActionErr('');
    if (a.key === 'cancel' && !window.confirm(`Batalkan pesanan ${o.id}? Kapasitas yang ditahan akan dilepas.`)) return;
    try {
      await api.orderAction(o.dbId, a.key);
      reload();
    } catch (e) {
      setActionErr(e.message);
    }
  }

  const nowLabel = now.toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' })
    + ' ' + now.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' }) + ' WIB';

  return (
    <div className="flex flex-col min-h-full">
      {cetak && <RekapCetak tanggal={cetak.tanggal} details={cetak.details} />}
      <div className="flex flex-col gap-6 p-9 flex-1 print:hidden">
        <div className="flex items-start justify-between">
          <div>
            <h1 className="text-[#2b1d16] text-[30px] font-normal">Daftar Pesanan</h1>
            <p className="text-[#8a7768] text-sm mt-1.5">Semua pesanan dari WhatsApp dan pembeli langsung, dalam satu catatan.</p>
          </div>
          <div className="flex gap-3">
            <button onClick={cetakPesanan} disabled={mencetak || produksiHariIni.length === 0}
              title={produksiHariIni.length === 0 ? 'Tidak ada pesanan yang diproduksi hari ini' : 'Rekap produk & jumlah box yang diproduksi hari ini'}
              className="btn-secondary text-[#4c0c13] disabled:opacity-50 disabled:cursor-not-allowed">
              <IconPrinter />
              {mencetak ? 'Menyiapkan…' : `Cetak pesanan hari ini (${produksiHariIni.length})`}
            </button>
            <button onClick={() => router.push('/input-pesanan')} className="btn-primary">+ Input Pesanan</button>
          </div>
        </div>

        <ErrorBanner message={error || actionErr} />

        <div className="card p-6 flex flex-col gap-5">
          <div className="flex gap-4 items-end">
            <div className="flex flex-col gap-2 flex-1 min-w-0">
              <label className="text-[#2b1d16] text-sm">Cari pesanan</label>
              <input className="input-field" placeholder="Cari kode atau nama pembeli" value={search} onChange={(e) => setSearch(e.target.value)} />
            </div>
            <div className="flex flex-col gap-2 w-[220px]">
              <label className="text-[#2b1d16] text-sm">Tanggal kirim</label>
              <input type="date" className="input-field" value={tanggal} onChange={(e) => setTanggal(e.target.value)} />
            </div>
            <div className="flex flex-col gap-2 w-[180px]">
              <label className="text-[#2b1d16] text-sm">Status</label>
              <select className="input-field" onChange={(e) => setTab(e.target.value)} value={activeTab}>
                {STATUS_TABS.map((s) => <option key={s} value={s}>{s}</option>)}
              </select>
            </div>
            <button className="btn-secondary h-11 text-[#4c0c13]" onClick={() => { setSearch(''); setTab('Semua'); setTanggal(''); }}>Atur ulang</button>
          </div>

          <div className="flex gap-3 flex-wrap">
            {STATUS_TABS.map((s) => (
              <button key={s} onClick={() => setTab(s)}
                className={`px-3 py-2 rounded-xl text-sm transition-colors ${activeTab === s ? 'bg-[#4c0c13] text-white font-semibold' : 'bg-[#faf6ee] text-[#8a7768] hover:bg-[#eadfd0]'}`}>
                {s} {tabCounts[s] ?? 0}
              </button>
            ))}
          </div>
        </div>

        <div className="card">
          <div className="flex items-center justify-between px-6 py-5 border-b border-[#eadfd0]">
            <div>
              <p className="text-[#2b1d16] text-lg font-normal">Semua pesanan</p>
              <p className="text-[#7e4722] text-xs mt-0.5">Status dihitung server · {nowLabel}</p>
            </div>
            <button onClick={reload} className="flex gap-2 items-center border border-[#eadfd0] rounded-lg px-3 h-9 text-sm text-[#2b1d16] font-semibold hover:bg-[#faf6ee]">
              <img src={iconArrowUpRight} alt="" className="size-4" />
              <span>Muat ulang</span>
            </button>
          </div>

          <div className="bg-[#faf6ee] flex items-center px-6 h-10 gap-4 border-b border-[#eadfd0]">
            <p className="w-[96px] shrink-0 text-xs font-medium text-[#7e4722]">Kode</p>
            <p className="flex-1 min-w-0 text-xs font-medium text-[#7e4722]">Pembeli</p>
            <p className="w-[300px] shrink-0 text-xs font-medium text-[#7e4722]">Produk</p>
            <p className="w-[60px] shrink-0 text-xs font-medium text-[#7e4722]">Qty box</p>
            <p className="w-[140px] shrink-0 text-xs font-medium text-[#7e4722]">Kirim / Produksi</p>
            <p className="w-[150px] shrink-0 text-xs font-medium text-[#7e4722]">Status</p>
            <p className="w-[36px] shrink-0 text-xs font-medium text-[#7e4722]">Aksi</p>
          </div>

          {loading && !orders && <div className="px-6 py-10 text-center text-[#8a7768] text-sm">Memuat pesanan…</div>}
          {!loading && filtered.length === 0 && (
            <div className="px-6 py-10 text-center text-[#8a7768] text-sm">Tidak ada pesanan untuk filter ini.</div>
          )}

          {filtered.map((o) => (
            <div key={o.id} className="flex items-center px-6 h-16 gap-4 border-b border-[#eadfd0] last:border-0 hover:bg-[#faf6ee]/50 transition-colors relative">
              <p className="w-[96px] text-sm font-semibold text-[#2b1d16]">{o.id}</p>
              <div className="flex-1 min-w-0">
                <p className="text-sm text-[#2b1d16] truncate">{o.pembeli}</p>
                {o.pj && <p className="text-xs text-[#8a7768] truncate" title={o.pjEmail || o.pj}>PJ: {o.pj}</p>}
              </div>
              <p className="w-[300px] shrink-0 text-sm font-medium text-[#4c0c13] truncate" title={o.produk} >
                {o.produk}
              </p>
              <p className="w-[60px] shrink-0 text-sm text-[#2b1d16]">
                {o.qty}
              </p>
              <div className="w-[140px] shrink-0">
                <p className="text-sm text-[#2b1d16]">
                  {fmtTgl(o.tanggalKirim)} · {o.jamKirim}
                </p>
                <p className="text-xs text-[#8a7768]">
                  {o.tanggalProduksi
                    ? `${fmtTgl(o.tanggalProduksi)} · ${hLabel(o.tanggalKirim, o.tanggalProduksi)}`
                    : 'Menunggu negosiasi'}
                </p>
              </div>
              <div className="w-[150px] shrink-0">
                <span className={badgeMap[o.statusOtomatis] ?? badgeMap['Diterima']}>
                  {statusText(o)}
                </span>
              </div>
              <div className="w-[36px] relative">
                <button onClick={() => setMenuOpen(menuOpen === o.id ? null : o.id)} className="text-[#8a7768] hover:text-[#2b1d16]">
                  <span className="text-base leading-none">···</span>
                </button>
                {menuOpen === o.id && (
                  <div className="absolute right-0 top-7 bg-white border border-[#eadfd0] rounded-xl shadow-modal z-10 min-w-[170px]">
                    {actionsFor(o).map((a) => (
                      <button key={a.key} onClick={() => runAction(o, a)}
                        className={`block w-full text-left px-4 py-2.5 text-sm hover:bg-[#faf6ee] first:rounded-t-xl ${a.danger ? 'text-[#dc2626]' : 'text-[#2b1d16]'}`}>
                        {a.label}
                      </button>
                    ))}
                    <button onClick={() => { setMenuOpen(null); setDetail(o); }}
                      className="block w-full text-left px-4 py-2.5 text-sm text-[#8a7768] hover:bg-[#faf6ee] rounded-b-xl border-t border-[#eadfd0]">
                      Lihat detail
                    </button>
                  </div>
                )}
              </div>
            </div>
          ))}

          <div className="px-6 py-4 border-t border-[#eadfd0]">
            <p className="text-[#7e4722] text-sm">Menampilkan {filtered.length} dari {list.length} pesanan</p>
          </div>
        </div>
      </div>

      <footer className="border-t border-[#eadfd0] bg-white px-9 py-3 flex items-center gap-2 print:hidden">
        <span className="text-[#8a7768] text-xs">ℹ</span>
        <p className="text-[#8a7768] text-xs">
          ≥72 jam sebelum kirim → Diterima · &lt;72 jam → Dijadwalkan · hari produksi + “Cetak pesanan hari ini” → Diproduksi · lewat jam kirim tanpa diselesaikan → Terlambat
        </p>
      </footer>
      <div className="bg-white border-t border-[#eadfd0] px-9 py-2 flex items-center justify-between print:hidden">
        <p className="text-[#8a7768] text-xs">Hakuu Bakery • Operasional usaha, bukan etalase pelanggan</p>
        <p className="text-[#8a7768] text-xs">Waktu Indonesia Barat</p>
      </div>

      {detail && <DetailModal order={detail} onClose={() => setDetail(null)} onChanged={reload} />}
    </div>
  );
}
