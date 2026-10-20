import { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import { iconIkonPengecekan } from '../../utils/icons';
import api from '../../services/bakery';
import { toUiProduct, variantsComplete, variantSummary } from '../../utils/mapper';
import { useLoad } from '../../utils/useLoad';
import ErrorBanner from '../../components/ErrorBanner';
import PackageCard from '../../components/pesanan/PackageCard';
import { fmtTgl, ymd } from '../../utils/orderStatus';
import { getNama, getEmail } from '../../utils/storage';

function CapacitySidebar({ items, preview, onPilihJam }) {
  const CAPACITY_DAILY = preview?.capacity ?? 800;
  const totalBox   = items.reduce((s, i) => s + i.qty, 0);
  // Beban kapasitas = box x bobot produk, dibulatkan ke atas (sama dengan perhitungan server).
  const totalBobot = Math.ceil(items.reduce((s, i) => s + i.qty * i.bobot, 0) - 1e-9);
  const totalUsed  = (preview?.used ?? 0) + totalBobot;
  const pct        = Math.round((totalUsed / CAPACITY_DAILY) * 100);
  const remaining  = CAPACITY_DAILY - totalUsed;

  const barColor = pct >= 100 ? '#b74b40' : pct >= 90 ? '#a66a17' : '#327a52';
  const label    = pct >= 100 ? 'Kapasitas penuh' : pct >= 90 ? 'Mendekati penuh' : 'Kapasitas cukup';
  const labelColor = pct >= 100 ? 'text-[#b74b40]' : pct >= 90 ? 'text-[#a66a17]' : 'text-[#327a52]';

  return (
    <div className="flex flex-col gap-4 w-[280px] shrink-0">
      {/* Capacity card */}
      <div className="card p-5 flex flex-col gap-4">
        <div className="flex items-start justify-between">
          <p className="text-[#2b1d16] text-sm font-semibold">Pengecekan kapasitas</p>
          <img src={iconIkonPengecekan} alt="" className="size-5" />
        </div>

        {preview?.dayLabel && (
          <p className="text-[#8a7768] text-xs -mt-2">Perkiraan produksi: {preview.dayLabel}</p>
        )}

        <div>
          <div className="flex justify-between mb-1.5">
            <span className="text-[#8a7768] text-xs">Kapasitas terpakai</span>
            <span className={`text-xs font-semibold ${labelColor}`}>{label}</span>
          </div>
          <div className="bg-[#eadfd0] rounded-full h-2 overflow-hidden">
            <div className="h-full rounded-full transition-all"
              style={{ width: `${Math.min(pct, 100)}%`, backgroundColor: barColor }} />
          </div>
          <div className="flex justify-between mt-1.5">
            <span className="text-[#8a7768] text-xs">{Math.round(totalUsed)} / {CAPACITY_DAILY} box</span>
            <span className={`text-xs font-semibold ${labelColor}`}>{pct}%</span>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-3">
          <div className="bg-[#faf6ee] rounded-xl p-3">
            <p className="text-[#8a7768] text-xs">Pesanan ini</p>
            <p className="text-[#2b1d16] text-base font-semibold mt-0.5">{totalBox} box</p>
            <p className="text-[#8a7768] text-[10px]">{totalBobot.toFixed(0)} box kapasitas</p>
          </div>
          <div className="bg-[#faf6ee] rounded-xl p-3">
            <p className="text-[#8a7768] text-xs">Sisa kapasitas</p>
            <p className={`text-base font-semibold mt-0.5 ${remaining < 0 ? 'text-[#b74b40]' : 'text-[#2b1d16]'}`}>
              {Math.max(0, Math.round(remaining))} box
            </p>
            <p className="text-[#8a7768] text-[10px]">setelah pesanan ini</p>
          </div>
        </div>

        {preview?.none === 'terlambat' && (
          <div className="bg-red-50 border border-red-200 rounded-xl p-3 text-xs text-[#b74b40]">
            ✕ Waktu terlalu dekat: pesanan tidak sempat diproduksi sebelum jam kirim.
          </div>
        )}
        {preview?.none === 'nego' && (
          <div className="bg-[#fdf6ec] border border-[#e8c98a] rounded-xl p-3 text-xs text-[#7e4722]">
            ⚠ Produksi H-1 sudah tidak memungkinkan. Pesanan bisa diproduksi di hari kirim bila jam kirim dinegosiasikan ke {preview.readyLabel} atau lebih.
          </div>
        )}
        {preview?.dialihkan && pct < 100 && (
          <div className="bg-[#fdf6ec] border border-[#e8c98a] rounded-xl p-3 text-xs text-[#7e4722]">
            ⚠ Kapasitas H-1 penuh. Produksi dialihkan ke hari kirim (jam kirim sudah {preview.readyLabel} atau lebih).
          </div>
        )}
        {!preview?.dialihkan && pct >= 90 && pct < 100 && (
          <div className="bg-[#fdf6ec] border border-[#e8c98a] rounded-xl p-3 text-xs text-[#7e4722]">
            ⚠ Kapasitas H-1 hampir penuh.
          </div>
        )}
        {pct >= 100 && (
          <div className="bg-red-50 border border-red-200 rounded-xl p-3 text-xs text-[#b74b40]">
            {preview?.bolehHariH
              ? '✕ Kapasitas H-1 dan hari kirim tidak cukup. Sistem akan mencoba menggeser pesanan lain; bila tetap penuh, pesanan ditolak.'
              : '✕ Kapasitas H-1 tidak cukup. Sistem akan mencoba menggeser pesanan lain terlebih dahulu.'}
          </div>
        )}
        {preview?.lewatBatas && (
          <div className="bg-red-50 border border-red-200 rounded-xl p-3 text-xs text-[#b74b40]">
            ✕ Jam kirim melewati batas pengiriman {preview.batasLabel}. Pesanan tidak bisa disimpan.
          </div>
        )}

        {/* Rekomendasi jam kirim: tetap di tanggal yang sama, mulai jam siap s.d. jam kirim terakhir */}
        {preview?.rekomendasi && (
          <div className="border border-[#eadfd0] rounded-xl p-3 flex flex-col gap-2">
            <p className="text-[#2b1d16] text-xs font-semibold">Rekomendasi jam kirim</p>
            {preview.rekomendasi.length > 0 ? (
              <>
                <p className="text-[#8a7768] text-xs">
                  Tanggal tetap sama, produksi dialihkan ke hari kirim. Pilih untuk mengganti jam kirim:
                </p>
                <div className="flex gap-2 flex-wrap">
                  {preview.rekomendasi.map((j) => (
                    <button key={j} type="button" onClick={() => onPilihJam(j)}
                      className="px-3 h-8 rounded-lg border border-[#4c0c13] text-[#4c0c13] text-xs font-semibold hover:bg-[#4c0c13] hover:text-white transition-colors">
                      {j.replace(':', '.')}
                    </button>
                  ))}
                </div>
              </>
            ) : (
              <p className="text-[#8a7768] text-xs">
                Tidak ada jam kirim yang masih muat di tanggal ini ({preview.readyLabel}–{preview.batasLabel}).
              </p>
            )}
          </div>
        )}
      </div>

      {/* Order summary */}
      {items.length > 0 && (
        <div className="card p-5 flex flex-col gap-3">
          <p className="text-[#2b1d16] text-sm font-semibold">Ringkasan pesanan</p>
          {items.map(item => (
            <div key={item.id} className="flex items-start justify-between gap-2">
              <div className="flex-1 min-w-0">
                <p className="text-[#2b1d16] text-xs font-medium truncate">{item.name}</p>
                <p className="text-[#8a7768] text-xs">×{item.qty} box</p>
                {item.maxVariants > 0 && (
                  <>
                    {variantSummary(item) && <p className="text-[#7e4722] text-xs">{variantSummary(item)}</p>}
                    {!variantsComplete(item) && (
                      <span className="inline-flex mt-1 px-2 py-0.5 rounded-lg text-[11px] font-medium bg-[#fdf6ec] text-[#a66a17] border border-[#e8c98a]">
                        Varian belum lengkap ({(item.chosenVariants || []).length}/{item.maxVariants})
                      </span>
                    )}
                  </>
                )}
              </div>
              <p className="text-[#2b1d16] text-xs font-semibold shrink-0">
                Rp {(item.price * item.qty).toLocaleString('id-ID')}
              </p>
            </div>
          ))}
          <div className="border-t border-[#eadfd0] pt-3 flex justify-between">
            <span className="text-[#2b1d16] text-sm font-semibold">Total</span>
            <span className="text-[#4c0c13] text-sm font-bold">
              Rp {items.reduce((s, i) => s + i.price * i.qty, 0).toLocaleString('id-ID')}
            </span>
          </div>
        </div>
      )}
    </div>
  );
}

export default function InputPesanan() {
  const router = useRouter();
  const [step, setStep]               = useState(1);
  const [pembeli, setPembeli]         = useState('');
  const [telepon, setTelepon]         = useState('');
  const [alamat, setAlamat]           = useState('');
  const [tanggalKirim, setTanggalKirim] = useState('');
  const [jamKirim, setJamKirim]       = useState('');
  const [catatan, setCatatan]         = useState('');
  const [activeMenuTab, setMenuTab]   = useState(null);
  const [orderItems, setOrderItems]   = useState([]);
  const [saving, setSaving]           = useState(false);
  const [error, setError]             = useState('');
  const [result, setResult]           = useState(null);   // hasil diterima / perlu_negosiasi
  const [rejected, setRejected]       = useState(null);   // hasil ditolak (422)
  const [preview, setPreview]         = useState(null);
  const [qtyDraft, setQtyDraft]       = useState({});     // teks qty yang sedang diketik (boleh kosong)

  const prods    = useLoad(() => api.getProducts().then(l => l.filter(p => p.aktif).map(toUiProduct)
    .map(p => ({ ...p, variants: p.variants.filter(v => v.active) }))), []);
  const cats     = useLoad(() => api.getCategories(), []);
  const settings = useLoad(() => api.getSettings(), []);

  const catList = cats.data || [];
  const activeCat = activeMenuTab && catList.includes(activeMenuTab) ? activeMenuTab : catList[0];
  const allProducts = (prods.data || []).filter(p => p.kategori === activeCat);
  // paket pilih varian tampil sebagai kartu; produk biasa tetap di tabel
  const packages = allProducts.filter(p => p.maxVariants > 0);
  const regularProducts = allProducts.filter(p => !p.maxVariants);
  const variantsIncomplete = orderItems.some(i => !variantsComplete(i));

  // Batas H-1 (jam siap, default 11.00): jam kirim sebelum ini wajib diproduksi H-1.
  // Hari H hanya boleh dipakai bila jam kirim >= jam siap.
  const prm = settings.data?.parameters || {};
  const prodEndHour = Number(prm.jam_produksi_selesai ?? 18);  // produksi 06.00–18.00
  const readyMin = Number(prm.batas_pagi_jam ?? 11) * 60;
  const readyLabel = `${String(Math.floor(readyMin / 60)).padStart(2, '0')}.${String(readyMin % 60).padStart(2, '0')}`;
  const lastHour = Number(prm.jam_kirim_terakhir ?? 20);       // jam kirim paling akhir (20.00)
  const batasLabel = `${String(lastHour).padStart(2, '0')}.00`;

  const jamValid = /^([01]\d|2[0-3]):[0-5]\d$/.test(jamKirim);
  const lewatBatas = jamValid && (Number(jamKirim.slice(0, 2)) * 60 + Number(jamKirim.slice(3, 5))) > lastHour * 60;
  // Langkah di indikator hanya bisa diklik bila tombol "Lanjut" menuju langkah itu aktif.
  const bisaLangkah2 = !!pembeli && !!tanggalKirim && jamValid && !lewatBatas;
  const bisaLangkah3 = bisaLangkah2 && orderItems.length > 0 && !variantsIncomplete;
  const bisaKeLangkah = { 1: true, 2: bisaLangkah2, 3: bisaLangkah3 };
  const maxRek = Number(prm.maks_tawaran_jam ?? 3);
  const stepJam = Math.max(1, Number(prm.pembulatan_jam ?? 1));
  const load = Math.ceil(orderItems.reduce((s, i) => s + i.qty * i.bobot, 0) - 1e-9);

  // Pratinjau kapasitas, meniru aturan server:
  //   H-1 dulu; hari H hanya bila jam kirim >= jam siap (11.00). Tidak ada H-2.
  useEffect(() => {
    if (!tanggalKirim || !jamValid) { setPreview(null); return; }
    let cancelled = false;
    const hariKirim = new Date(`${tanggalKirim}T00:00:00`);
    const [hh, mm] = jamKirim.split(':').map(Number);
    const bolehHariH = hh * 60 + mm >= readyMin;
    const h1 = new Date(hariKirim); h1.setDate(h1.getDate() - 1);
    const last = bolehHariH ? hariKirim : h1;
    // produksi hari ini sudah selesai -> paling cepat besok
    const now = new Date();
    const today = new Date(now); today.setHours(0, 0, 0, 0);
    if (now.getHours() >= prodEndHour) today.setDate(today.getDate() + 1);
    const from = h1 < today ? today : h1;
    if (from > hariKirim) { setPreview({ none: 'terlambat' }); return; }  // hari H pun sudah lewat
    const nego = from > last;                                             // H-1 lewat, jam kirim < 11.00
    // Ambil H-1 (bila masih bisa) s.d. hari H: hari H dipakai untuk rekomendasi jam kirim.
    const days = Math.round((hariKirim - from) / 86400000) + 1;
    api.getSchedule(ymd(from), days)
      .then(d => { if (!cancelled) setPreview({ allDays: d, days: d.filter(x => x.date <= ymd(last)), bolehHariH, tanggalKirim, nego }); })
      .catch(() => { if (!cancelled) setPreview(nego ? { none: 'nego' } : null); });
    return () => { cancelled = true; };
  }, [tanggalKirim, jamKirim, jamValid, readyMin, prodEndHour]);

  /*
     Rekomendasi jam kirim (meniru server): tanggal tetap sama, hanya jam yang lebih lambat,
     paling awal jam siap (11.00), paling akhir jam kirim terakhir (20.00), dibulatkan per jam.
     Jam valid bila H-1 masih muat, atau (jam >= 11.00 dan hari H masih muat).
  */
  function hitungRekomendasi() {
    if (!preview?.allDays || !jamValid || load <= 0) return [];
    const freeOf = (date) => {
      const d = preview.allDays.find(x => x.date === date);
      return d ? d.capacity - d.used : -1;
    };
    const h1 = new Date(`${tanggalKirim}T00:00:00`); h1.setDate(h1.getDate() - 1);
    const h1Free = freeOf(ymd(h1));
    const hFree = freeOf(tanggalKirim);
    const [hh, mm] = jamKirim.split(':').map(Number);
    const req = hh * 60 + mm;
    const step = stepJam * 60;
    let t = Math.ceil(req / step) * step;
    if (t <= req) t += step;
    if (t < readyMin) t = readyMin;
    const out = [];
    for (; t <= lastHour * 60 && out.length < maxRek; t += step) {
      if (h1Free >= load || (t >= readyMin && hFree >= load)) {
        out.push(`${String(Math.floor(t / 60)).padStart(2, '0')}:${String(t % 60).padStart(2, '0')}`);
      }
    }
    return out;
  }

  // Hari pertama yang masih muat (meniru urutan server); bila tidak ada, tampilkan hari terakhir.
  let sidebarPreview = null;
  if (preview?.none) sidebarPreview = { none: preview.none, readyLabel };
  else if (preview?.nego) {
    // H-1 sudah lewat: langsung tampilkan rekomendasi jam di hari kirim
    sidebarPreview = { none: 'nego', readyLabel, batasLabel, lewatBatas, rekomendasi: load > 0 ? hitungRekomendasi() : null };
  } else if (preview?.days?.length) {
    const fit = preview.days.find(d => d.capacity - d.used >= load);
    const day = fit || preview.days[0];   // tidak muat -> tampilkan H-1 (yang dicek pertama)
    const hariH = day.date === preview.tanggalKirim;
    sidebarPreview = {
      capacity: day.capacity, used: day.used,
      dayLabel: `${fmtTgl(day.date)} (${hariH ? 'hari H' : 'H-1'})`,
      readyLabel, batasLabel, lewatBatas, bolehHariH: preview.bolehHariH, dialihkan: hariH,
      // rekomendasi hanya muncul bila jam yang diminta tidak muat
      rekomendasi: !fit && load > 0 ? hitungRekomendasi() : null,
    };
  }

  function addItem(p) {
    const ex = orderItems.find(i => i.id === p.id);
    if (ex) setOrderItems(orderItems.map(i => i.id === p.id ? { ...i, qty: i.qty + 1 } : i));
    else    setOrderItems([...orderItems, { ...p, qty: 1 }]);
  }
  function updateQty(id, qty) {
    if (qty <= 0) setOrderItems(orderItems.filter(i => i.id !== id));
    else          setOrderItems(orderItems.map(i => i.id === id ? { ...i, qty } : i));
  }
  // paket: qty 0 menghapus baris (pilihan varian ikut di-reset)
  function setPackageQty(p, qty) {
    if (qty <= 0) setOrderItems(orderItems.filter(i => i.id !== p.id));
    else if (orderItems.some(i => i.id === p.id)) setOrderItems(orderItems.map(i => i.id === p.id ? { ...i, qty } : i));
    else setOrderItems([...orderItems, { ...p, qty, chosenVariants: [] }]);
  }
  function setPackageVariants(id, chosenVariants) {
    setOrderItems(orderItems.map(i => i.id === id ? { ...i, chosenVariants } : i));
  }
  function clearDraft(id) {
    setQtyDraft(d => { const n = { ...d }; delete n[id]; return n; });
  }
  // dipanggil saat user mengetik di kotak qty: kotak boleh kosong sementara
  function handleQtyTyping(id, raw) {
    const text = raw.replace(/[^0-9]/g, '');
    setQtyDraft(d => ({ ...d, [id]: text }));
    const val = parseInt(text, 10);
    if (!isNaN(val) && val > 0) updateQty(id, val);
  }
  // dipanggil saat kotak qty ditinggalkan: kalau kosong / 0, kembali ke 1
  function commitQty(id) {
    const val = parseInt(qtyDraft[id], 10);
    if (isNaN(val) || val < 1) updateQty(id, 1);
    clearDraft(id);
  }
  function stepQty(id, qty) {
    clearDraft(id);
    updateQty(id, qty);
  }

  async function handleSimpan() {
    setSaving(true);
    setError('');
    setRejected(null);
    try {
      const res = await api.insertOrder({
        customer: { nama: pembeli.trim(), telepon: telepon.trim(), alamat: alamat.trim() },
        items: orderItems.map(i => ({
          product_id: i.id,
          qty: i.qty,
          ...(i.maxVariants > 0 && {
            variants: (i.chosenVariants || []).map(v => ({ variant_id: v.id, ...(v.opsi && { opsi: v.opsi }) })),
          }),
        })),
        deliver_at: `${tanggalKirim} ${jamKirim}`,
        note: catatan.trim(),
      });
      setResult(res);
    } catch (e) {
      if (e.status === 422 && e.data?.outcome === 'ditolak') setRejected(e.data);
      else setError(e.message);
    } finally {
      setSaving(false);
    }
  }

  function applyAlternative(alt) {
    setTanggalKirim(alt.slice(0, 10));
    setJamKirim(alt.slice(11, 16));
    setRejected(null);
    setStep(1);
  }

  function resetForm() {
    setResult(null); setStep(1); setPembeli(''); setTelepon(''); setAlamat('');
    setTanggalKirim(''); setJamKirim(''); setCatatan(''); setOrderItems([]);
  }

  if (result) {
    const nego = result.outcome === 'perlu_negosiasi';
    return (
      <div className="flex items-center justify-center h-full p-9">
        <div className="card p-10 text-center max-w-lg flex flex-col gap-4">
          <div className={`rounded-full size-16 flex items-center justify-center mx-auto ${nego ? 'bg-[#fdf6ec]' : 'bg-green-100'}`}>
            <span className={`text-3xl ${nego ? 'text-[#a66a17]' : 'text-green-600'}`}>{nego ? '!' : '✓'}</span>
          </div>
          <h2 className="text-[#2b1d16] text-xl font-semibold">
            {nego ? `Pesanan ${result.kode || ''} perlu negosiasi jam kirim` : `Pesanan ${result.kode} Tersimpan!`}
          </h2>
          <p className="text-[#8a7768] text-xs -mt-2">Penanggung jawab: {getNama() || getEmail() || '—'}</p>
          {!nego && (
            <p className="text-[#8a7768] text-sm">
              Dijadwalkan produksi {fmtTgl(result.production_day)}.
              {result.overtime_pcs > 0 && ` Memakai lembur ${result.overtime_pcs} box kapasitas.`}
            </p>
          )}
          {!nego && result.moved?.length > 0 && (
            <div className="bg-[#fdf6ec] border border-[#e8c98a] rounded-xl p-3 text-left text-xs text-[#7e4722] flex flex-col gap-1">
              <p className="font-semibold">Pesanan lain digeser agar muat:</p>
              {result.moved.map(m => <p key={m.order_id}>{m.kode}: {fmtTgl(m.from)} → {fmtTgl(m.to)}</p>)}
            </div>
          )}
          {nego && (
            <>
              <p className="text-[#8a7768] text-sm">Kapasitas pada jam yang diminta tidak cukup. Alternatif jam kirim yang direkomendasikan:</p>
              <div className="flex flex-col gap-1.5">
                {(result.offers || []).map(o => (
                  <p key={o.id} className="text-sm text-[#2b1d16] bg-[#faf6ee] rounded-lg py-2">
                    {fmtTgl(o.deliver_at)} · {o.deliver_at.slice(11, 16)} <span className="text-[#8a7768] text-xs">(produksi {fmtTgl(o.production_day)})</span>
                  </p>
                ))}
              </div>
              <p className="text-[#8a7768] text-xs">Buka detail pesanan di Daftar Pesanan untuk menawarkannya ke pembeli.</p>
            </>
          )}
          <div className="flex gap-3 justify-center pt-2">
            <button onClick={resetForm} className="btn-secondary">Input pesanan lain</button>
            <button onClick={() => router.push('/daftar-pesanan')} className="btn-primary">Lihat daftar pesanan</button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-6 p-9">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-[#2b1d16] text-[30px] font-normal">Input Pesanan</h1>
          <p className="text-[#8a7768] text-sm mt-1">
            Catat pesanan baru dan cek ruang produksi sebelum menyepakati pengiriman.
          </p>
        </div>
        <div className="bg-[#faf6ee] border border-[#eadfd0] px-3 py-1.5 rounded-full">
          <span className="text-[#4c0c13] text-xs font-medium">Draf pesanan baru</span>
        </div>
      </div>

      {/* Step indicator */}
      <div className="bg-[#faf6ee] border border-[#eadfd0] rounded-xl px-5 py-3.5 flex items-center gap-7">
        {[
          { n: 1, label: '1  Detail pesanan' },
          { n: 2, label: '2  Pengecekan kapasitas' },
          { n: 3, label: '3  Simpan pesanan' },
        ].map((s, i) => (
          <div key={s.n} className="flex items-center gap-7">
            <button
              onClick={() => setStep(s.n)}
              disabled={!bisaKeLangkah[s.n]}
              className={`text-sm font-semibold whitespace-pre transition-colors disabled:cursor-not-allowed disabled:opacity-50 ${
                step >= s.n ? 'text-[#4c0c13]' : 'text-[#8a7768] font-normal'
              }`}
            >
              {s.label}
            </button>
            {i < 2 && <span className="text-[#8a7768] text-base">›</span>}
          </div>
        ))}
      </div>

      {/* Two-column layout */}
      <div className="flex gap-6 items-start">
        {/* Left: main form */}
        <div className="flex flex-col gap-5 flex-1 min-w-0">

          {/* Step 1: pembeli info */}
          {(step === 1 || step === 2 || step === 3) && (
            <div className="card p-6 flex flex-col gap-5">
              <div className="flex items-start justify-between">
                <p className="text-[#2b1d16] text-lg font-semibold">Informasi pembeli & pengiriman</p>
                {/* tombol Edit hanya di langkah 2; di langkah Simpan pesanan tidak ditampilkan */}
                {step === 2 && (
                  <button onClick={() => setStep(1)} className="text-[#4c0c13] text-xs font-semibold hover:underline">
                    Edit
                  </button>
                )}
              </div>

              {step === 1 && (
                <>
                  <div className="grid grid-cols-2 gap-4">
                    <div className="flex flex-col gap-2">
                      <label className="text-[#2b1d16] text-sm">Nama pembeli *</label>
                      <input className="input-field" value={pembeli}
                        onChange={e => setPembeli(e.target.value)} placeholder="Nurul Hidayah" />
                    </div>
                    <div className="flex flex-col gap-2">
                      <label className="text-[#2b1d16] text-sm">Nomor telepon *</label>
                      <input className="input-field" value={telepon}
                        onChange={e => setTelepon(e.target.value)} placeholder="+62 812-3456-7890" />
                    </div>
                  </div>
                  <div className="flex flex-col gap-2">
                    <label className="text-[#2b1d16] text-sm">Alamat pengiriman *</label>
                    <input className="input-field" value={alamat}
                      onChange={e => setAlamat(e.target.value)}
                      placeholder="Jl. Raya Pajajaran No. 12, Kota Bogor" />
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div className="flex flex-col gap-2">
                      <label className="text-[#2b1d16] text-sm">Tanggal kirim *</label>
                      <input type="date" className="input-field" min={ymd()} value={tanggalKirim}
                        onChange={e => setTanggalKirim(e.target.value)} />
                    </div>
                    <div className="flex flex-col gap-2">
                      <label className="text-[#2b1d16] text-sm">Jam kirim *</label>
                      <input
                        type="text"
                        className={`input-field font-mono tracking-wider ${
                          (jamKirim && !/^([01]\d|2[0-3]):[0-5]\d$/.test(jamKirim)) || lewatBatas
                            ? 'border-[#dc2626] focus:border-[#dc2626]'
                            : ''
                        }`}
                        value={jamKirim}
                        onChange={e => {
                          // auto-insert colon after 2 digits
                          let v = e.target.value.replace(/[^0-9:]/g, '');
                          if (v.length === 2 && !v.includes(':') && jamKirim.length < 2) {
                            v = v + ':';
                          }
                          if (v.length <= 5) setJamKirim(v);
                        }}
                        onBlur={e => {
                          const v = e.target.value;
                          // normalise single-digit hour: "9:30" → "09:30"
                          const match = v.match(/^(\d):(\d{2})$/);
                          if (match) setJamKirim(`0${match[1]}:${match[2]}`);
                        }}
                        placeholder="09:00"
                        maxLength={5}
                      />
                      {jamKirim && !/^([01]\d|2[0-3]):[0-5]\d$/.test(jamKirim) && (
                        <p className="text-[#dc2626] text-xs">
                          Format 24 jam: 00:00 – 23:59
                        </p>
                      )}
                      {lewatBatas && (
                        <p className="text-[#dc2626] text-xs">
                          Jam kirim paling lambat {batasLabel}
                        </p>
                      )}
                    </div>
                  </div>
                  <div className="flex flex-col gap-2">
                    <label className="text-[#2b1d16] text-sm">Catatan (opsional)</label>
                    <textarea className="input-field min-h-[80px] resize-none" value={catatan}
                      onChange={e => setCatatan(e.target.value)}
                      placeholder="Permintaan khusus, alergi, dll." />
                  </div>
                  <button
                    onClick={() => setStep(2)}
                    disabled={!bisaLangkah2}
                    className="btn-primary self-end disabled:opacity-50 disabled:cursor-not-allowed">
                    Lanjut ke Pilih Produk →
                  </button>
                </>
              )}

              {step > 1 && (
                <div className="grid grid-cols-2 gap-x-6 gap-y-2 text-sm">
                  <div><span className="text-[#8a7768]">Pembeli</span> <span className="text-[#2b1d16] ml-2 font-medium">{pembeli}</span></div>
                  <div><span className="text-[#8a7768]">Telepon</span> <span className="text-[#2b1d16] ml-2">{telepon}</span></div>
                  <div><span className="text-[#8a7768]">Kirim</span> <span className="text-[#2b1d16] ml-2">{tanggalKirim} · {jamKirim}</span></div>
                  <div><span className="text-[#8a7768]">Penanggung jawab</span> <span className="text-[#2b1d16] ml-2">{getNama() || getEmail() || '—'}</span></div>
                </div>
              )}
            </div>
          )}

          {/* Step 2: choose products */}
          {step >= 2 && (
            <div className="card p-6 flex flex-col gap-5">
              <div className="flex items-start justify-between">
                <p className="text-[#2b1d16] text-lg font-semibold">Pilih produk</p>
              </div>

              {step === 2 && (
                <>
                  {/* Menu tabs */}
                  <div className="flex gap-3 flex-wrap">
                    {catList.map(c => (
                      <button key={c} onClick={() => setMenuTab(c)}
                        className={`px-4 h-9 rounded-lg text-sm font-semibold transition-colors ${
                          activeCat === c
                            ? 'bg-[#4c0c13] text-white'
                            : 'bg-white border border-[#eadfd0] text-[#2b1d16] hover:bg-[#faf6ee]'
                        }`}>
                        {c}
                      </button>
                    ))}
                  </div>

                  {/* Paket pilih varian */}
                  {packages.length > 0 && (
                    <div className="flex flex-col gap-3">
                      {packages.map(p => (
                        <PackageCard key={p.id} pkg={p}
                          item={orderItems.find(i => i.id === p.id)}
                          onQty={qty => setPackageQty(p, qty)}
                          onVariants={v => setPackageVariants(p.id, v)} />
                      ))}
                    </div>
                  )}

                  {/* Product list */}
                  {(regularProducts.length > 0 || packages.length === 0) && (
                  <div className="flex flex-col gap-0">
                    {/* Table header */}
                    <div className="bg-[#faf6ee] flex items-center px-4 h-10 gap-4 rounded-t-xl border border-[#eadfd0]">
                      <p className="flex-1 min-w-0 text-xs font-medium text-[#7e4722]">Nama produk</p>
                      <p className="w-[120px] text-xs font-medium text-[#7e4722]">Harga / box</p>
                      <p className="w-[120px] text-xs font-medium text-[#7e4722]">Jenis</p>
                      <p className="w-[120px] text-xs font-medium text-[#7e4722] text-center">Jumlah</p>
                    </div>
                    {prods.loading && !prods.data && <p className="text-[#8a7768] text-sm py-4">Memuat produk…</p>}
                    {!prods.loading && regularProducts.length === 0 && <p className="text-[#8a7768] text-sm py-4">Belum ada produk aktif di kategori ini.</p>}
                    {regularProducts.map((p, idx) => {
                      const inOrder = orderItems.find(i => i.id === p.id);
                      return (
                        <div key={p.id}
                          className={`flex items-center px-4 h-14 gap-4 border-l border-r border-b border-[#eadfd0] ${
                            idx === regularProducts.length - 1 ? 'rounded-b-xl' : ''
                          } hover:bg-[#faf6ee]/50 transition-colors`}>
                          <p className="flex-1 min-w-0 text-sm font-semibold text-[#2b1d16]">{p.name}</p>
                          <p className="w-[120px] text-sm text-[#2b1d16]">
                            Rp {p.price.toLocaleString('id-ID')}
                          </p>
                          <p className="w-[120px] text-sm text-[#78685c]">{p.jenis}</p>
                          <div className="w-[120px] flex items-center justify-center">
                            {inOrder ? (
                              <div className="flex items-center gap-1.5">
                                <button
                                  onClick={() => stepQty(p.id, inOrder.qty - 1)}
                                  className="size-7 rounded-full border border-[#eadfd0] text-[#4c0c13] font-bold text-lg leading-none hover:bg-[#faf6ee] shrink-0"
                                >
                                  −
                                </button>
                                <input
                                  type="text"
                                  inputMode="numeric"
                                  aria-label={`Jumlah box ${p.name}`}
                                  value={qtyDraft[p.id] ?? String(inOrder.qty)}
                                  onFocus={e => e.target.select()}
                                  onChange={e => handleQtyTyping(p.id, e.target.value)}
                                  onBlur={() => commitQty(p.id)}
                                  onKeyDown={e => { if (e.key === 'Enter') e.currentTarget.blur(); }}
                                  className="w-12 text-center text-sm font-semibold text-[#2b1d16] border border-[#eadfd0] rounded-lg py-1 focus:outline-none focus:border-[#4c0c13] [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
                                />
                                <button
                                  onClick={() => stepQty(p.id, inOrder.qty + 1)}
                                  className="size-7 rounded-full bg-[#4c0c13] text-white font-bold text-lg leading-none hover:bg-[#3a0a0f] shrink-0"
                                >
                                  +
                                </button>
                              </div>
                            ) : (
                              <button onClick={() => addItem(p)}
                                className="size-7 rounded-full bg-[#4c0c13] text-white font-bold text-lg leading-none hover:bg-[#3a0a0f]">
                                +
                              </button>
                            )}
                          </div>
                        </div>
                      );
                    })}
                  </div>
                  )}

                  {variantsIncomplete && (
                    <p className="text-[#a66a17] text-xs">Lengkapi pilihan varian paket sebelum melanjutkan.</p>
                  )}
                  <div className="flex gap-3 pt-2">
                    <button onClick={() => setStep(1)} className="btn-secondary">← Kembali</button>
                    <button onClick={() => setStep(3)} disabled={!bisaLangkah3}
                      className="btn-primary disabled:opacity-50 disabled:cursor-not-allowed">
                      Lanjut ke Konfirmasi →
                    </button>
                  </div>
                </>
              )}

              {step === 3 && orderItems.length > 0 && (
                <div className="flex flex-col gap-2 text-sm">
                  {orderItems.map(i => (
                    <div key={i.id} className="flex justify-between">
                      <span className="text-[#2b1d16]">
                        {i.name} ×{i.qty}
                        {i.maxVariants > 0 && <span className="block text-xs text-[#7e4722]">{variantSummary(i) || 'Varian belum dipilih'}</span>}
                      </span>
                      <span className="text-[#2b1d16] font-medium">
                        Rp {(i.price * i.qty).toLocaleString('id-ID')}
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* Step 3: confirm & save */}
          {step === 3 && (
            <div className="card p-6 flex flex-col gap-4">
              <p className="text-[#2b1d16] text-lg font-semibold">Simpan pesanan</p>
              <p className="text-[#8a7768] text-sm">
                Periksa kembali detail pesanan sebelum disimpan.
              </p>
              <ErrorBanner message={error} />
              {rejected && (
                <div className="bg-red-50 border border-red-200 rounded-xl p-4 flex flex-col gap-2">
                  <p className="text-[#b74b40] text-sm font-semibold">Pesanan ditolak: {rejected.reason}</p>
                  {rejected.alternative_dates?.length > 0 && (
                    <>
                      <p className="text-[#8a7768] text-xs">Jam kirim alternatif yang masih muat:</p>
                      <div className="flex gap-2 flex-wrap">
                        {rejected.alternative_dates.map(a => (
                          <button key={a} onClick={() => applyAlternative(a)} className="btn-secondary text-[#4c0c13] text-xs h-9 px-3">
                            {fmtTgl(a)} · {a.slice(11, 16)}
                          </button>
                        ))}
                      </div>
                    </>
                  )}
                </div>
              )}
              <div className="flex gap-3 pt-2">
                <button onClick={() => setStep(2)} className="btn-secondary">← Kembali</button>
                <button onClick={handleSimpan} disabled={saving || variantsIncomplete || orderItems.length === 0 || lewatBatas} className="btn-primary flex-1 justify-center disabled:opacity-60">
                  {saving ? 'Menyimpan…' : 'Simpan Pesanan'}
                </button>
              </div>
            </div>
          )}
        </div>

        {/* Right: capacity sidebar */}
        <CapacitySidebar items={orderItems} preview={sidebarPreview} onPilihJam={setJamKirim} />
      </div>
    </div>
  );
}
