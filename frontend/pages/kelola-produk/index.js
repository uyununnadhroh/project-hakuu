import { useState } from 'react';
import api from '../../services/bakery';
import { JENIS_VALUE, hitungBobot, komposisiProblem, toUiProduct } from '../../utils/mapper';
import { useLoad } from '../../utils/useLoad';
import ErrorBanner from '../../components/ErrorBanner';
import VariantEditor, { toEditorVariant, toVariantPayload, variantProblem } from '../../components/produk/VariantEditor';

// Toggle image URLs from Figma
const imgToggleOn  = "/assets/img-toggle-on.svg";
const imgToggleOff = "/assets/img-toggle-off.svg";

const PAKET_KATEGORI = 'Paket Pilih Varian';

function ProductModal({ mode, product, cats, onClose, onSave }) {
  const [form, setForm] = useState(() => {
    const base = product
      ? { ...product }
      : { name: '', jenis: 'Tanpa topping', price: '', active: true, kategori: cats[0] || '', isiBox: 3, isiTopping: 0, maxVariants: 0, variants: [] };
    return {
      ...base,
      isiBox: String(base.isiBox),
      isiTopping: String(base.jenis === 'Campuran' ? base.isiTopping : 1),
      isPaket: base.maxVariants > 0 || (!product && base.kategori === PAKET_KATEGORI),
      maxVariants: base.maxVariants > 0 ? String(base.maxVariants) : '',
      variants: base.variants.map(toEditorVariant),
    };
  });
  const paketProblem = form.isPaket ? variantProblem(form.maxVariants, form.variants) : '';
  const isi = Number(form.isiBox);
  const topping = Number(form.isiTopping);
  const isiProblem = komposisiProblem(form.jenis, isi, topping);
  const bobot = isiProblem ? null : hitungBobot(form.jenis, isi, topping);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [err, setErr] = useState('');
  const [saving, setSaving] = useState(false);

  async function submit(f) {
    setErr('');
    setSaving(true);
    try {
      await onSave(f);
    } catch (e) {
      setErr(e.message);
      setSaving(false);
    }
  }

  return (
    <div className="fixed inset-0 bg-black/40 flex items-start justify-end z-50">
      <div className="bg-white h-full w-[520px] flex flex-col shadow-modal overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-6 border-b border-[#eadfd0]">
          <div>
            <h2 className="text-[#2b1d16] text-xl font-semibold">
              {mode === 'add' ? 'Tambah Produk' : 'Edit Produk'}
            </h2>
            <p className="text-[#8a7768] text-sm mt-0.5">
              {mode === 'add' ? 'Isi detail produk baru.' : 'Ubah detail produk yang ada.'}
            </p>
          </div>
          <button onClick={onClose} className="text-[#8a7768] text-xl leading-none hover:text-[#2b1d16]">×</button>
        </div>

        {/* Body */}
        <div className="flex flex-col gap-5 p-6 overflow-y-auto flex-1">
          <div className="flex flex-col gap-2">
            <label className="text-[#2b1d16] text-sm">Nama produk</label>
            <input className="input-field" value={form.name}
              onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
              placeholder="Mis. Roti Sobek Abon" />
          </div>
          <div className="flex flex-col gap-2">
            <label className="text-[#2b1d16] text-sm">Kategori</label>
            <select className="input-field" value={form.kategori}
              onChange={e => {
                const kategori = e.target.value;
                // kategori paket otomatis menyalakan pengaturan varian
                setForm(f => ({ ...f, kategori, isPaket: f.isPaket || kategori === PAKET_KATEGORI }));
              }}>
              {cats.map(c => <option key={c}>{c}</option>)}
            </select>
          </div>
          <div className="flex flex-col gap-2">
            <label className="text-[#2b1d16] text-sm">Jenis produksi</label>
            <select className="input-field" value={form.jenis}
              onChange={e => setForm(f => ({ ...f, jenis: e.target.value }))}>
              <option>Tanpa topping</option>
              <option>Bertopping</option>
              <option>Campuran</option>
            </select>
          </div>
          <div className="flex gap-3">
            <div className="flex flex-col gap-2 flex-1">
              <label className="text-[#2b1d16] text-sm">Isi per box (roti)</label>
              <input className="input-field" type="number" min="1" value={form.isiBox}
                onChange={e => setForm(f => ({ ...f, isiBox: e.target.value }))}
                placeholder="3" />
            </div>
            {form.jenis === 'Campuran' && (
              <div className="flex flex-col gap-2 flex-1">
                <label className="text-[#2b1d16] text-sm">Roti bertopping per box</label>
                <input className="input-field" type="number" min="1" value={form.isiTopping}
                  onChange={e => setForm(f => ({ ...f, isiTopping: e.target.value }))}
                  placeholder="1" />
              </div>
            )}
          </div>
          {isiProblem
            ? <p className="text-[#b74b40] text-xs -mt-3">{isiProblem}</p>
            : (
              <div className="bg-[#faf6ee] rounded-xl px-4 py-3 -mt-2">
                <p className="text-[#2b1d16] text-sm">
                  Bobot kapasitas: <span className="font-semibold">{bobot.toLocaleString('id-ID', { minimumFractionDigits: 2 })} ×</span>
                </p>
                <p className="text-[#8a7768] text-xs mt-0.5">
                  100 box produk = {Math.ceil(100 * bobot - 1e-9)} box kapasitas
                </p>
              </div>
            )}
          <div className="flex flex-col gap-2">
            <label className="text-[#2b1d16] text-sm">Harga (Rp)</label>
            <input className="input-field" type="number" value={form.price}
              onChange={e => setForm(f => ({ ...f, price: e.target.value }))}
              placeholder="18000" />
          </div>
          <div className="flex flex-col gap-2">
            <label className="text-[#2b1d16] text-sm">Tipe menu</label>
            <div className="flex gap-3">
              {[false, true].map(v => (
                <button key={String(v)} type="button"
                  onClick={() => setForm(f => ({ ...f, isPaket: v }))}
                  className={`flex-1 h-10 rounded-xl border text-sm transition-colors ${
                    form.isPaket === v
                      ? 'border-[#4c0c13] bg-[#4c0c13]/5 text-[#4c0c13] font-semibold'
                      : 'border-[#eadfd0] text-[#8a7768]'
                  }`}>
                  {v ? 'Paket pilih varian' : 'Produk biasa'}
                </button>
              ))}
            </div>
            {form.isPaket && (
              <VariantEditor
                maxVariants={form.maxVariants}
                variants={form.variants}
                onMaxChange={maxVariants => setForm(f => ({ ...f, maxVariants }))}
                onChange={variants => setForm(f => ({ ...f, variants }))}
              />
            )}
          </div>
          <div className="flex flex-col gap-2">
            <label className="text-[#2b1d16] text-sm">Status aktif</label>
            <div className="flex gap-3">
              {[true, false].map(v => (
                <button key={String(v)} type="button"
                  onClick={() => setForm(f => ({ ...f, active: v }))}
                  className={`flex-1 h-10 rounded-xl border text-sm transition-colors ${
                    form.active === v
                      ? 'border-[#4c0c13] bg-[#4c0c13]/5 text-[#4c0c13] font-semibold'
                      : 'border-[#eadfd0] text-[#8a7768]'
                  }`}>
                  {v ? 'Aktif' : 'Nonaktif'}
                </button>
              ))}
            </div>
          </div>

          {mode === 'edit' && !showDeleteConfirm && (
            <button type="button" onClick={() => setShowDeleteConfirm(true)}
              className="text-[#dc2626] text-sm hover:underline text-left mt-2">
              Hapus produk ini
            </button>
          )}
          <ErrorBanner message={err} />
          {showDeleteConfirm && (
            <div className="bg-red-50 border border-red-200 rounded-xl p-4 flex flex-col gap-3 mt-2">
              <p className="text-[#dc2626] text-sm font-semibold">Hapus produk?</p>
              <p className="text-[#8a7768] text-xs">Tindakan ini tidak dapat dibatalkan. Produk yang sudah pernah dipesan tidak bisa dihapus, nonaktifkan saja.</p>
              <div className="flex gap-2">
                <button onClick={() => submit({ ...form, deleted: true })}
                  className="flex-1 bg-[#dc2626] text-white text-sm h-9 rounded-lg hover:bg-red-700">
                  Ya, hapus
                </button>
                <button onClick={() => setShowDeleteConfirm(false)}
                  className="flex-1 border border-[#eadfd0] text-[#2b1d16] text-sm h-9 rounded-lg">
                  Batal
                </button>
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="flex gap-3 px-6 py-5 border-t border-[#eadfd0]">
          <button onClick={onClose} className="btn-secondary flex-1 justify-center">Batal</button>
          <button disabled={saving || !form.name?.trim() || form.price === '' || !!paketProblem || !!isiProblem} onClick={() => submit(form)} className="btn-primary flex-1 justify-center disabled:opacity-60">
            {mode === 'add' ? 'Tambah Produk' : 'Simpan Perubahan'}
          </button>
        </div>
      </div>
    </div>
  );
}

function KategoriPanel({ cats, onClose, onChanged }) {
  const [newCat, setNewCat] = useState('');
  const [deleteTarget, setDeleteTarget] = useState(null);
  const [err, setErr] = useState('');

  async function addCat() {
    if (!newCat.trim()) return;
    setErr('');
    try { await api.insertCategory(newCat.trim()); setNewCat(''); onChanged(); } catch (e) { setErr(e.message); }
  }
  async function removeCat(cat) {
    setErr('');
    try { await api.deleteCategory(cat); setDeleteTarget(null); onChanged(); } catch (e) { setErr(e.message); setDeleteTarget(null); }
  }

  return (
    <div className="fixed inset-0 bg-black/40 flex items-start justify-end z-50">
      <div className="bg-white h-full w-[520px] flex flex-col shadow-modal">
        <div className="flex items-center justify-between px-6 py-6 border-b border-[#eadfd0]">
          <div>
            <h2 className="text-[#2b1d16] text-xl font-semibold">Kelola Kategori</h2>
            <p className="text-[#8a7768] text-sm mt-0.5">Atur kategori menu produk.</p>
          </div>
          <button onClick={onClose} className="text-[#8a7768] text-xl hover:text-[#2b1d16]">×</button>
        </div>
        <div className="flex flex-col gap-0 flex-1 overflow-y-auto">
          {cats.map((cat, i) => (
            <div key={i}>
              <div className="flex items-center justify-between px-6 py-4 border-b border-[#eadfd0]">
                <span className="text-[#2b1d16] text-sm">{cat}</span>
                <button onClick={() => setDeleteTarget(cat)}
                  className="text-[#dc2626] text-xs hover:underline">Hapus</button>
              </div>
              {deleteTarget === cat && (
                <div className="mx-6 mb-4 mt-2 bg-red-50 border border-red-200 rounded-xl p-3 flex flex-col gap-2">
                  <p className="text-[#dc2626] text-xs font-semibold">
                    Kategori tidak dapat dihapus jika masih ada produk di dalamnya.
                  </p>
                  <div className="flex gap-2">
                    <button onClick={() => removeCat(cat)}
                      className="flex-1 bg-[#dc2626] text-white text-xs h-8 rounded-lg">Ya, hapus</button>
                    <button onClick={() => setDeleteTarget(null)}
                      className="flex-1 border border-[#eadfd0] text-[#2b1d16] text-xs h-8 rounded-lg">Batal</button>
                  </div>
                </div>
              )}
            </div>
          ))}
          {err && <div className="mx-6 mt-4 bg-red-50 border border-red-200 rounded-xl px-4 py-3 text-sm text-[#b74b40]">{err}</div>}
          <div className="flex gap-2 p-6">
            <input className="input-field" value={newCat}
              onChange={e => setNewCat(e.target.value)}
              placeholder="Nama kategori baru"
              onKeyDown={e => { if (e.key === 'Enter') addCat(); }} />
            <button onClick={addCat}
              className="btn-primary shrink-0">Tambah</button>
          </div>
        </div>
        <div className="px-6 py-5 border-t border-[#eadfd0]">
          <button onClick={onClose} className="btn-secondary w-full justify-center">Selesai</button>
        </div>
      </div>
    </div>
  );
}

export default function KelolaProduk() {
  const prods = useLoad(() => api.getProducts().then(l => l.map(toUiProduct)), []);
  const catsLoad = useLoad(() => api.getCategories(), []);
  const [tab, setTab] = useState(null);
  const [search, setSearch] = useState('');
  const [filterJenis, setFilterJenis] = useState('Semua jenis');
  const [filterStatus, setFilterStatus] = useState('Semua status');
  const [modal, setModal] = useState(null); // null | 'add' | 'edit' | 'kategori'
  const [selectedProduct, setSelectedProduct] = useState(null);
  const [err, setErr] = useState('');

  const catList = catsLoad.data || [];
  const activeTab = tab && catList.includes(tab) ? tab : catList[0];
  const all = prods.data || [];
  const products = all.filter(p => p.kategori === activeTab);

  const filtered = products.filter(p => {
    const matchSearch = p.name.toLowerCase().includes(search.toLowerCase());
    const matchJenis  = filterJenis === 'Semua jenis' || p.jenis === filterJenis;
    const matchStatus = filterStatus === 'Semua status'
      || (filterStatus === 'Aktif' && p.active)
      || (filterStatus === 'Nonaktif' && !p.active);
    return matchSearch && matchJenis && matchStatus;
  });

  async function handleToggle(p) {
    setErr('');
    try { await api.updateProduct(p.id, { aktif: !p.active }); prods.reload(); } catch (e) { setErr(e.message); }
  }

  // Melempar galat agar modal tetap terbuka dan menampilkan pesannya.
  async function handleSave(form) {
    // produk biasa: isi paket 0 & tanpa varian (varian lama dihapus/dinonaktifkan server)
    const paket = {
      maks_varian: form.isPaket ? Number(form.maxVariants) : 0,
      variants: form.isPaket ? form.variants.map(toVariantPayload) : [],
    };
    // bobot dihitung server dari jenis + isi box
    const komposisi = {
      jenis: JENIS_VALUE[form.jenis],
      isi_box: Number(form.isiBox),
      isi_topping: form.jenis === 'Campuran' ? Number(form.isiTopping) : undefined,
    };
    if (form.deleted) {
      await api.deleteProduct(form.id);
    } else if (modal === 'add') {
      await api.insertProduct({ nama: form.name.trim(), harga: Number(form.price), kategori: form.kategori, ...komposisi, ...paket });
    } else {
      await api.updateProduct(form.id, {
        nama: form.name.trim(), harga: Number(form.price), aktif: form.active, kategori: form.kategori,
        ...komposisi, ...paket,
      });
    }
    prods.reload();
    setModal(null);
    setSelectedProduct(null);
  }

  return (
    <div className="flex flex-col gap-6 p-9">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-[#2b1d16] text-[30px] font-normal">Kelola Produk</h1>
          <p className="text-[#8a7768] text-sm mt-1">
            Atur produk dan bobotnya agar pengecekan kapasitas tetap akurat.
          </p>
        </div>
        <div className="flex gap-3">
          <button onClick={() => setModal('kategori')} className="btn-secondary text-[#4c0c13]">
            Kelola Kategori
          </button>
          <button onClick={() => setModal('add')} className="btn-primary">
            + Tambah Produk
          </button>
        </div>
      </div>

      <ErrorBanner message={err || prods.error || catsLoad.error} />

      {/* Weight rule cards */}
      <div className="grid grid-cols-3 gap-5">
        <div className="card p-6 flex flex-col gap-2.5">
          <div className="flex items-start justify-between">
            <p className="text-[#2b1d16] text-lg font-semibold">Tanpa topping</p>
            <span className="bg-[#faf6ee] text-[#4c0c13] text-xs font-medium px-3 py-1.5 rounded-full">
              Bobot dasar
            </span>
          </div>
          <p className="text-[#2b1d16] text-[28px] font-semibold">1,00 ×</p>
          <p className="text-[#8a7768] text-sm">Per box, berapa pun isinya</p>
        </div>
        <div className="card p-6 flex flex-col gap-2.5">
          <div className="flex items-start justify-between">
            <p className="text-[#2b1d16] text-lg font-semibold">Bertopping</p>
            <span className="bg-[#faf6ee] text-[#4c0c13] text-xs font-medium px-3 py-1.5 rounded-full">
              Patokan isi 3
            </span>
          </div>
          <p className="text-[#2b1d16] text-[28px] font-semibold">0,41 × isi</p>
          <p className="text-[#8a7768] text-sm">Isi 3 = 1,23 · isi 2 = 0,82 · isi 4 = 1,64</p>
        </div>
        <div className="card p-6 flex flex-col gap-2.5">
          <div className="flex items-start justify-between">
            <p className="text-[#2b1d16] text-lg font-semibold">Campuran</p>
            <span className="bg-[#faf6ee] text-[#4c0c13] text-xs font-medium px-3 py-1.5 rounded-full">
              Dihitung per roti
            </span>
          </div>
          <p className="text-[#2b1d16] text-[22px] font-semibold leading-[42px]">0,41 × topping + sisa ÷ isi</p>
          <p className="text-[#8a7768] text-sm">Isi 3 (1 topping + 2 polos) = 1,08</p>
        </div>
      </div>

      {/* Category tabs */}
      <div className="flex gap-3 flex-wrap">
        {catList.map(c => (
          <button key={c} onClick={() => setTab(c)}
            className={`px-4 h-10 rounded-lg text-sm font-semibold transition-colors ${
              activeTab === c
                ? 'bg-[#4c0c13] text-white'
                : 'bg-white border border-[#eadfd0] text-[#2b1d16] hover:bg-[#faf6ee]'
            }`}>
            {c} ({all.filter(p => p.kategori === c).length})
          </button>
        ))}
      </div>

      {/* Filter bar */}
      <div className="flex gap-4 items-end">
        <div className="flex flex-col gap-2 flex-1">
          <label className="text-[#2b1d16] text-sm">Cari produk</label>
          <input className="input-field" placeholder="Cari nama produk"
            value={search} onChange={e => setSearch(e.target.value)} />
        </div>
        <div className="flex flex-col gap-2 w-[200px]">
          <label className="text-[#2b1d16] text-sm">Jenis</label>
          <select className="input-field" value={filterJenis} onChange={e => setFilterJenis(e.target.value)}>
            <option>Semua jenis</option>
            <option>Tanpa topping</option>
            <option>Bertopping</option>
            <option>Campuran</option>
          </select>
        </div>
        <div className="flex flex-col gap-2 w-[200px]">
          <label className="text-[#2b1d16] text-sm">Status</label>
          <select className="input-field" value={filterStatus} onChange={e => setFilterStatus(e.target.value)}>
            <option>Semua status</option>
            <option>Aktif</option>
            <option>Nonaktif</option>
          </select>
        </div>
      </div>

      {/* Table */}
      <div className="card">
        {/* Table toolbar */}
        <div className="flex items-center justify-between px-6 py-6 border-b border-[#eadfd0]">
          <div>
            <p className="text-[#2b1d16] text-lg font-semibold">
              {activeTab}
            </p>
            <p className="text-[#7e4722] text-xs mt-1">
              {filtered.length} produk
            </p>
          </div>
        </div>

        {/* Table header */}
        <div className="bg-[#faf6ee] flex items-center px-6 h-10 gap-4 border-b border-[#eadfd0]">
          <p className="flex-1 min-w-0 text-xs font-medium text-[#7e4722]">Nama produk</p>
          <p className="w-[120px] text-xs font-medium text-[#7e4722]">Harga</p>
          <p className="w-[180px] text-xs font-medium text-[#7e4722]">Jenis produksi</p>
          <p className="w-[120px] text-xs font-medium text-[#7e4722]">Status aktif</p>
          <p className="w-[80px] text-xs font-medium text-[#7e4722]">Aksi</p>
        </div>

        {filtered.map((p) => (
          <div key={p.id} className="flex items-center px-6 h-16 gap-4 border-b border-[#eadfd0] last:border-0 hover:bg-[#faf6ee]/50 transition-colors">
            <div className="flex-1 min-w-0">
              <p className="text-sm font-semibold text-[#2b1d16] truncate">{p.name}</p>
              {p.maxVariants > 0 && (
                <p className="text-xs text-[#7e4722] truncate" title={p.variants.filter(v => v.active).map(v => v.name).join(', ')}>
                  Isi {p.maxVariants} varian · {p.variants.filter(v => v.active).length} varian aktif
                </p>
              )}
            </div>
            <p className="w-[120px] text-sm text-[#2b1d16]">Rp {p.price.toLocaleString('id-ID')}</p>
            <div className="w-[180px]">
              <p className="text-sm text-[#78685c]">{p.jenis}</p>
              <p className="text-xs text-[#8a7768]">
                Isi {p.isiBox}{p.jenis === 'Campuran' ? ` (${p.isiTopping} topping)` : ''} · bobot {Number(p.bobot).toLocaleString('id-ID', { minimumFractionDigits: 2 })}
              </p>
            </div>
            <div className="w-[120px]">
              <button onClick={() => handleToggle(p)} className="block">
                <img
                  src={p.active ? imgToggleOn : imgToggleOff}
                  alt={p.active ? 'Aktif' : 'Nonaktif'}
                  className="h-6 w-11"
                />
              </button>
            </div>
            <div className="w-[80px]">
              <button
                onClick={() => { setSelectedProduct(p); setModal('edit'); }}
                className="text-sm font-semibold text-[#4c0c13] hover:underline"
              >
                Edit
              </button>
            </div>
          </div>
        ))}

        {prods.loading && !prods.data && <div className="px-6 py-8 text-center text-[#8a7768] text-sm">Memuat produk…</div>}
        {!prods.loading && filtered.length === 0 && <div className="px-6 py-8 text-center text-[#8a7768] text-sm">Belum ada produk di kategori ini.</div>}
      </div>

      {/* Modals */}
      {(modal === 'add' || modal === 'edit') && (
        <ProductModal
          mode={modal}
          product={modal === 'edit' ? selectedProduct : null}
          cats={catList}
          onClose={() => { setModal(null); setSelectedProduct(null); }}
          onSave={handleSave}
        />
      )}
      {modal === 'kategori' && (
        <KategoriPanel cats={catList} onClose={() => setModal(null)} onChanged={() => { catsLoad.reload(); prods.reload(); }} />
      )}
    </div>
  );
}
