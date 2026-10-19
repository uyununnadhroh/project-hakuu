import { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import { useDispatch } from 'react-redux';
import apiCore from '../services/core';
import { getToken, setSession } from '../utils/storage';
import { setUser } from '../redux/actions/userAction';
import {
  logoHakuuBakeryLogin,
  iconClipboardListLogin,
  iconGauge,
  iconCalendarCheck,
} from '../utils/icons';

function IconEye() {
  return (
    <svg viewBox="0 0 24 24" className="size-5" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M2.06 12.35a1 1 0 0 1 0-.7 10.75 10.75 0 0 1 19.88 0 1 1 0 0 1 0 .7 10.75 10.75 0 0 1-19.88 0" />
      <circle cx="12" cy="12" r="3" />
    </svg>
  );
}

function IconEyeOff() {
  return (
    <svg viewBox="0 0 24 24" className="size-5" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M10.73 5.08A10.43 10.43 0 0 1 12 5c4.97 0 8.4 3.6 9.94 6.65a1 1 0 0 1 0 .7 10.7 10.7 0 0 1-1.44 2.49" />
      <path d="M14.08 14.16a3 3 0 0 1-4.24-4.24" />
      <path d="M17.48 17.5a10.75 10.75 0 0 1-15.42-5.15 1 1 0 0 1 0-.7 10.75 10.75 0 0 1 4.45-5.14" />
      <path d="m2 2 20 20" />
    </svg>
  );
}

export default function Login() {
  const router = useRouter();
  const dispatch = useDispatch();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  // sudah login -> langsung ke dashboard
  useEffect(() => {
    if (getToken()) router.replace('/dashboard');
  }, []);

  if (getToken()) return null;

  async function handleSubmit(e) {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      const r = await apiCore.login(email.trim(), password);
      // selalu tetap login sampai tombol Keluar ditekan (sama seperti default "Ingat saya" sebelumnya)
      setSession(r.token, r.nama, true, r.email || email.trim());
      dispatch(setUser({ nama: r.nama || '' }));
      router.push('/dashboard');
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex items-stretch min-h-screen bg-[#faf6ee]">
      {/* Left panel */}
      <div className="bg-[#4c0c13] flex flex-col items-start justify-between p-16 w-[600px] shrink-0">
        {/* Logo */}
        <div>
          <img src={logoHakuuBakeryLogin} alt="Hakuu Bakery" className="h-[50px] w-[183px] object-contain" />
        </div>

        {/* Benefits */}
        <div className="flex flex-col gap-8 w-full">
          <div>
            <p className="text-white text-5xl font-semibold leading-[1.2]">Pesanan tertata.</p>
            <p className="text-white text-5xl font-semibold leading-[1.2]">Produksi lebih tenang.</p>
          </div>
          <p className="text-[#fcebc8] text-base leading-relaxed">
            Satu ruang kerja untuk mencatat pesanan, menjaga kapasitas, dan menyiapkan roti tepat waktu.
          </p>
          <div className="flex flex-col gap-5">
            {[
              {
                icon: iconClipboardListLogin,
                title: 'Catat semua pesanan',
                desc: 'WhatsApp dan pesanan langsung, satu tempat.',
              },
              {
                icon: iconGauge,
                title: 'Cek kapasitas sebelum menerima',
                desc: 'Hitung beban produksi, bukan sekadar jumlah box.',
              },
              {
                icon: iconCalendarCheck,
                title: 'Susun prioritas produksi',
                desc: 'Jadwal H-1 dan hari H yang mudah dipantau.',
              },
            ].map((item) => (
              <div key={item.title} className="flex gap-4 items-center">
                <div className="bg-white/10 flex items-center justify-center rounded-xl size-11 shrink-0">
                  <img src={item.icon} alt="" className="size-5" />
                </div>
                <div className="flex flex-col gap-1">
                  <p className="text-white text-sm font-semibold">{item.title}</p>
                  <p className="text-[#fcebc8] text-xs">{item.desc}</p>
                </div>
              </div>
            ))}
          </div>
        </div>

        <p className="text-[#daa278] text-xs">Dibuat untuk operasional UMKM Hakuu Bakery.</p>
      </div>

      {/* Right panel */}
      <div className="flex flex-1 flex-col items-center justify-center gap-6 p-20">
        <div className="card p-10 w-[460px] flex flex-col gap-7">
          {/* Header */}
          <div className="flex flex-col gap-2.5">
            <h1 className="text-[#2b1d16] text-3xl font-bold">Selamat datang kembali</h1>
            <p className="text-[#8a7768] text-sm">Masuk untuk menyiapkan hari produksi Anda.</p>
          </div>

          {/* Form */}
          <form onSubmit={handleSubmit} className="flex flex-col gap-7">
            {/* Email */}
            <div className="flex flex-col gap-2">
              <label className="text-[#2b1d16] text-sm">Email</label>
              <input
                type="email"
                className="input-field"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="admin@hakuubakery.com"
              />
            </div>

            {/* Password */}
            <div className="flex flex-col gap-2">
              <label className="text-[#2b1d16] text-sm">Password</label>
              <div className="relative">
                <input
                  type={showPassword ? 'text' : 'password'}
                  className="input-field pr-12"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="••••••••••••"
                />
                <button
                  type="button"
                  className="absolute right-2 top-1/2 -translate-y-1/2 flex items-center justify-center size-8 rounded-lg text-[#8a7768] hover:text-[#4c0c13] hover:bg-[#faf6ee] focus:outline-none focus-visible:ring-2 focus-visible:ring-[#4c0c13]/40 transition-colors"
                  onClick={() => setShowPassword(!showPassword)}
                  aria-label={showPassword ? 'Sembunyikan password' : 'Tampilkan password'}
                  title={showPassword ? 'Sembunyikan password' : 'Tampilkan password'}
                >
                  {showPassword ? <IconEyeOff /> : <IconEye />}
                </button>
              </div>
            </div>

            {error && (
              <div className="bg-red-50 border border-red-200 rounded-xl px-4 py-3 text-sm text-[#b74b40]">{error}</div>
            )}

            <button type="submit" disabled={loading || !email || !password}
              className="btn-primary justify-center h-11 w-full rounded-xl disabled:opacity-60 disabled:cursor-not-allowed">
              {loading ? 'Memproses…' : 'Masuk'}
            </button>
          </form>
        </div>

        <p className="text-[#8a7768] text-xs">©2026 Hakuu Bakery</p>
      </div>
    </div>
  );
}
