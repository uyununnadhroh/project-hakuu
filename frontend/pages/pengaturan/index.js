import { useState, useEffect } from 'react';
import api from '../../services/bakery';
import apiCore from '../../services/core';
import { getEmail } from '../../utils/storage';
import { useLoad } from '../../utils/useLoad';
import ErrorBanner from '../../components/ErrorBanner';
import { ymd } from '../../utils/orderStatus';

function NotifToggle({ label, defaultOn }) {
  const [on, setOn] = useState(defaultOn);
  return (
    <div className="flex items-center justify-between">
      <span className="text-[#2b1d16] text-sm">{label}</span>
      <button
        type="button"
        onClick={() => setOn(!on)}
        className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${on ? 'bg-[#4c0c13]' : 'bg-[#eadfd0]'}`}
      >
        <span className={`inline-block size-4 rounded-full bg-white shadow transition-transform ${on ? 'translate-x-6' : 'translate-x-1'}`} />
      </button>
    </div>
  );
}

const initials = (s) => (s || '?').replace(/@.*/, '').split(/[\s._-]+/).filter(Boolean).slice(0, 2).map((w) => w[0].toUpperCase()).join('') || '?';

/** Kartu daftar email yang terdaftar & bisa mengakses website */
function AkunTerdaftar() {
  const emailSaya = getEmail().toLowerCase();
  const { data, error, loading, reload } = useLoad(async () => {
    try {
      const r = await apiCore.getUsers();
      return { users: Array.isArray(r) ? r : (r?.users ?? []), belumAda: false };
    } catch (e) {
      // cadangan bila backend versi lama belum punya endpoint /api/users
      if (e.status === 404 || e.status === 405) return { users: [], belumAda: true };
      throw e;
    }
  }, []);

  const users = [...(data?.users ?? [])].sort((a, b) =>
    (b.email || '').toLowerCase() === emailSaya ? 1 : (a.email || '').toLowerCase() === emailSaya ? -1 : (a.email || '').localeCompare(b.email || ''));

  return (
    <div className="card">
      <div className="flex items-start justify-between gap-4 px-6 py-5 border-b border-[#eadfd0]">
        <div>
          <h2 className="text-[#2b1d16] text-base font-semibold">Akun terdaftar</h2>
          <p className="text-[#8a7768] text-xs mt-1">Email yang bisa masuk ke website Hakuu Bakery Planner.</p>
        </div>
        <div className="flex items-center gap-3 shrink-0">
          {data && !data.belumAda && (
            <span className="bg-[#faf6ee] text-[#4c0c13] text-xs font-medium px-3 py-1.5 rounded-full">{users.length} akun</span>
          )}
          <button type="button" onClick={reload} className="border border-[#eadfd0] rounded-lg px-3 h-9 text-sm text-[#2b1d16] font-semibold hover:bg-[#faf6ee]">
            Muat ulang
          </button>
        </div>
      </div>

      {loading && !data && <p className="px-6 py-6 text-[#8a7768] text-sm">Memuat daftar akun…</p>}
      {error && <div className="px-6 py-4"><ErrorBanner message={error} /></div>}

      {data?.belumAda && (
        <div className="px-6 py-5 flex flex-col gap-3">
          <div className="bg-[#fdf6ec] border border-[#e8c98a] rounded-xl px-4 py-3 text-sm text-[#a66a17]">
            Daftar akun belum bisa ditampilkan karena server belum menyediakan data akun (<code>/api/users</code>).
          </div>
          {emailSaya && (
            <div className="flex items-center gap-3">
              <div className="bg-[#daa278] flex items-center justify-center rounded-lg size-9 shrink-0">
                <span className="text-[#4c0c13] text-xs font-semibold">{initials(emailSaya)}</span>
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm text-[#2b1d16] truncate">{emailSaya}</p>
                <p className="text-xs text-[#8a7768]">Akun yang sedang dipakai</p>
              </div>
            </div>
          )}
        </div>
      )}

      {data && !data.belumAda && users.length === 0 && (
        <p className="px-6 py-6 text-[#8a7768] text-sm">Belum ada akun terdaftar.</p>
      )}

      {users.map((u, idx) => {
        const saya = (u.email || '').toLowerCase() === emailSaya;
        return (
          <div key={u.id ?? u.email} className={`flex items-center gap-3 px-6 py-3.5 ${idx > 0 ? 'border-t border-[#eadfd0]' : ''}`}>
            <div className={`flex items-center justify-center rounded-lg size-9 shrink-0 ${saya ? 'bg-[#4c0c13]' : 'bg-[#daa278]'}`}>
              <span className={`text-xs font-semibold ${saya ? 'text-[#fcebc8]' : 'text-[#4c0c13]'}`}>{initials(u.nama || u.email)}</span>
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-semibold text-[#2b1d16] truncate">{u.email}</p>
              <p className="text-xs text-[#8a7768] truncate">
                {u.nama || 'Tanpa nama'}
                {u.created_at ? ` · terdaftar ${new Date(u.created_at).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })}` : ''}
              </p>
            </div>
            {saya && <span className="bg-[#fcebc8] text-[#7e4722] text-xs font-medium px-2.5 py-1 rounded-lg shrink-0">Sedang login</span>}
          </div>
        );
      })}
    </div>
  );
}

export default function Pengaturan() {
  const [capacity, setCapacity] = useState(800);
  const [shopName, setShopName] = useState('Hakuu Bakery');
  const [email, setEmail] = useState('admin@hakuubakery.com');
  const [saved, setSaved] = useState(false);
  const [error, setError] = useState('');
  const [saving, setSaving] = useState(false);

  const settings = useLoad(() => api.getSettings(), []);
  const today = useLoad(() => api.getSchedule(ymd(), 1).then((d) => d[0]), []);
  const used = today.data?.used ?? 0;

  useEffect(() => {
    const c = settings.data?.parameters?.kapasitas_harian;
    if (c) setCapacity(c);
  }, [settings.data]);

  // Hanya kapasitas harian yang disimpan ke server; sisanya belum punya endpoint.
  async function handleSave(e) {
    e.preventDefault();
    setError('');
    setSaving(true);
    try {
      await api.updateSettings({ kapasitas_harian: Number(capacity) });
      settings.reload();
      today.reload();
      setSaved(true);
      setTimeout(() => setSaved(false), 3000);
    } catch (err) {
      setError(err.message);
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="flex flex-col gap-6 p-9 max-w-2xl">
      <div>
        <h1 className="text-[#2b1d16] text-3xl font-normal">Pengaturan</h1>
        <p className="text-[#8a7768] text-sm mt-1.5">Kelola konfigurasi toko dan sistem.</p>
      </div>

      <form onSubmit={handleSave} className="flex flex-col gap-6">
        {/* Shop info */}
        <div className="card p-6 flex flex-col gap-5">
          <h2 className="text-[#2b1d16] text-base font-semibold">Informasi Toko</h2>
          <p className="text-[#8a7768] text-xs -mt-3">Belum disimpan ke server (hanya tampilan).</p>
          <div className="flex flex-col gap-2">
            <label className="text-[#2b1d16] text-sm">Nama toko</label>
            <input
              className="input-field"
              value={shopName}
              onChange={(e) => setShopName(e.target.value)}
            />
          </div>
          <div className="flex flex-col gap-2">
            <label className="text-[#2b1d16] text-sm">Email admin</label>
            <input
              type="email"
              className="input-field"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
            />
          </div>
        </div>

        {/* Akun yang bisa mengakses website */}
        <AkunTerdaftar />

        {/* Production settings */}
        <div className="card p-6 flex flex-col gap-5">
          <h2 className="text-[#2b1d16] text-base font-semibold">Kapasitas Produksi</h2>
          <div className="flex flex-col gap-2">
            <label className="text-[#2b1d16] text-sm">
              Kapasitas harian (satuan bobot)
            </label>
            <div className="flex items-center gap-3">
              <input
                type="number"
                className="input-field max-w-[120px]"
                value={capacity}
                onChange={(e) => setCapacity(Number(e.target.value))}
                min={100}
                step={50}
              />
              <span className="text-[#8a7768] text-sm">unit per hari</span>
            </div>
            <p className="text-[#8a7768] text-xs">
              Kapasitas ini digunakan untuk menghitung persentase beban produksi.
            </p>
          </div>
          <div className="bg-[#faf6ee] border border-[#eadfd0] rounded-xl p-4">
            <div className="flex justify-between mb-2">
              <span className="text-[#2b1d16] text-sm">Kapasitas saat ini</span>
              <span className="text-[#4c0c13] text-sm font-semibold">{capacity} unit</span>
            </div>
            <div className="bg-[#eadfd0] rounded-full h-2 overflow-hidden">
              <div className="bg-[#4c0c13] h-full rounded-full" style={{ width: `${Math.min(100, Math.round(used / capacity * 100))}%` }} />
            </div>
            <p className="text-[#8a7768] text-xs mt-2">
              Hari ini: {used} / {capacity} unit ({Math.round(used / capacity * 100)}% terpakai)
            </p>
          </div>
        </div>

        {/* Notification settings */}
        <div className="card p-6 flex flex-col gap-5">
          <h2 className="text-[#2b1d16] text-base font-semibold">Notifikasi</h2>
          <p className="text-[#8a7768] text-xs -mt-3">Belum disimpan ke server (hanya tampilan).</p>
          {[
            { label: 'Peringatan kapasitas hampir penuh (≥80%)', defaultOn: true },
            { label: 'Pesanan baru masuk', defaultOn: true },
            { label: 'Pesanan menunggu konfirmasi > 2 jam', defaultOn: false },
          ].map((n) => <NotifToggle key={n.label} label={n.label} defaultOn={n.defaultOn} />)}
        </div>

        <ErrorBanner message={error || settings.error} />

        {/* Save */}
        <div className="flex items-center gap-4">
          <button type="submit" disabled={saving} className="btn-primary px-8 disabled:opacity-60">
            {saving ? 'Menyimpan…' : 'Simpan Pengaturan'}
          </button>
          {saved && (
            <span className="text-[#16a34a] text-sm font-medium flex items-center gap-1">
              ✓ Tersimpan
            </span>
          )}
        </div>
      </form>
    </div>
  );
}
