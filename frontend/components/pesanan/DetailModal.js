import { useState } from 'react';
import api from '../../services/bakery';
import { penanggungJawab } from '../../utils/mapper';
import { useLoad } from '../../utils/useLoad';
import ErrorBanner from '../ErrorBanner';
import ScoreBadge from '../ScoreBadge';
import { badgeMap, fmtTgl, hLabel, statusText } from '../../utils/orderStatus';

const rp = (n) => `Rp ${Number(n).toLocaleString('id-ID')}`;

const OFFER_LABEL = {
  direkomendasikan: 'Rekomendasi',
  ditawarkan: 'Menunggu jawaban pembeli',
  diterima: 'Diterima pembeli',
  ditolak: 'Ditolak pembeli',
  kedaluwarsa: 'Kedaluwarsa',
};

const KRITERIA = [
  { key: 'J', label: 'Urgensi jam kirim (J)' },
  { key: 'Q', label: 'Jumlah pesanan (Q)' },
  { key: 'W', label: 'Waktu pesanan masuk (W)' },
];
// Status yang skornya dihitung ulang tiap jam (sama dengan GetOrdersForRescore di backend).
const STATUS_DISKOR = ['baru', 'terjadwal', 'menunggu_negosiasi', 'at_risk'];
const fmtBobot = (n) => Number(n).toLocaleString('id-ID', { minimumFractionDigits: 3, maximumFractionDigits: 3 });

/** Skor prioritas AHP: total + rincian J/Q/W (skala 0–100) beserta bobotnya. */
function SkorPrioritas({ d }) {
  // bobot dari pengaturan; bila gagal dimuat, rincian tetap tampil tanpa bobot
  const { data: settings } = useLoad(() => api.getSettings(), []);
  const w = settings?.weights;
  const parts = d.score_parts;
  const dihitung = d.score != null && parts && STATUS_DISKOR.includes(d.status);

  return (
    <div className="bg-[#faf6ee] border border-[#eadfd0] rounded-xl px-4 py-3 flex flex-col gap-2">
      <div className="flex items-center justify-between">
        <p className="text-[#2b1d16] text-sm font-semibold">Skor prioritas</p>
        {dihitung && <ScoreBadge score={d.score} />}
      </div>
      {dihitung ? (
        <>
          <div className="flex flex-col gap-1">
            {KRITERIA.map((k) => (
              <p key={k.key} className="text-xs text-[#7e4722]">
                {k.label} <span className="font-semibold text-[#2b1d16]">{Math.round((parts[k.key] ?? 0) * 100)}</span>
                {w?.[k.key] != null && <span className="text-[#8a7768]"> · bobot {fmtBobot(w[k.key])}</span>}
              </p>
            ))}
          </div>
          <p className="text-[#8a7768] text-xs">
            Skor dihitung ulang otomatis setiap jam; makin tinggi skor, makin didahulukan pada hari produksinya.
          </p>
        </>
      ) : (
        <p className="text-[#8a7768] text-xs">Skor tidak dihitung untuk status ini</p>
      )}
    </div>
  );
}

export default function DetailModal({ order, onClose, onChanged }) {
  const { data: d, error, loading, reload } = useLoad(() => api.getOrder(order.dbId), [order.dbId]);
  const pj = d ? penanggungJawab(d) : null;
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState('');

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

              <SkorPrioritas d={d} />

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
