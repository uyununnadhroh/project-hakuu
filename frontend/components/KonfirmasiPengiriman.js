import { useCallback, useEffect, useState } from "react";
import api from "../services/bakery";

/*
   note : konfirmasi pengiriman untuk penjual (dipasang di Navbar)
          - batch pagi dibuka 11.00 (kirim hari ini < 11.00), batch malam 20.00 (kirim >= 11.00)
          - per pesanan: "Selesai" (final) atau "Belum terkirim" (+ alasan opsional, masih bisa jadi Selesai)
          - jendela muncul sendiri sekali tiap ada batch baru yang belum dikonfirmasi
          - tombol notifikasi tetap tampil selama masih ada pesanan yang belum Selesai
          - jam & batch dihitung server dalam WIB
*/
const POLL_MS = 60000;
const SEEN_KEY = "hakuu_konfirmasi_dilihat";

const jam = (s) => (s ? s.slice(11, 16) : "");
const tgl = (s) => new Date(`${s.slice(0, 10)}T00:00:00`).toLocaleDateString("id-ID", { day: "numeric", month: "short" });

function readSeen() {
    try { return localStorage.getItem(SEEN_KEY) || ""; } catch { return ""; }
}
function writeSeen(v) {
    try { localStorage.setItem(SEEN_KEY, v); } catch { /* penyimpanan browser tidak tersedia */ }
}

function StatusChip({ k }) {
    if (!k) return <span className="text-[#8a7768] text-xs">Belum dikonfirmasi</span>;
    const oleh = [jam(k.waktu) && `${tgl(k.waktu)} ${jam(k.waktu)}`, k.oleh && `oleh ${k.oleh}`].filter(Boolean).join(" ");
    if (k.status === "selesai") {
        return <span className="text-xs text-[#166534]"><b className="font-semibold">Selesai</b>{oleh && ` · ${oleh}`}</span>;
    }
    return (
        <span className="text-xs text-[#a66a17]">
            <b className="font-semibold">Belum terkirim</b>{k.alasan && ` · ${k.alasan}`}{oleh && ` · ${oleh}`}
        </span>
    );
}

function OrderRow({ o, today, onDone }) {
    const [mode, setMode] = useState(null); // null | "alasan"
    const [alasan, setAlasan] = useState(o.konfirmasi?.alasan || "");
    const [busy, setBusy] = useState(false);
    const [err, setErr] = useState("");
    const selesai = o.konfirmasi?.status === "selesai";

    async function kirim(status) {
        if (status === "selesai" && !window.confirm(`Tandai ${o.kode} sudah sampai ke pelanggan? Setelah Selesai tidak bisa diubah lagi.`)) return;
        setBusy(true);
        setErr("");
        try {
            await api.confirmDelivery(o.id, status, status === "belum_terkirim" ? alasan.trim() : "");
            setMode(null);
            await onDone();
        } catch (e) {
            setErr(e.message);
        } finally {
            setBusy(false);
        }
    }

    return (
        <div className="px-5 py-3.5 border-t border-[#eadfd0] flex flex-col gap-2">
            <div className="flex items-start gap-3">
                <div className="flex-1 min-w-0">
                    <p className="text-[#2b1d16] text-sm">
                        <span className="font-semibold">{o.kode}</span> · {o.customer || "—"}
                    </p>
                    <p className="text-[#7e4722] text-xs truncate" title={o.products}>{o.products}</p>
                    <StatusChip k={o.konfirmasi} />
                </div>
                <p className="text-[#2b1d16] text-sm whitespace-nowrap">
                    Kirim {o.deliver_at.slice(0, 10) !== today && `${tgl(o.deliver_at)} `}{jam(o.deliver_at)}
                </p>
            </div>
            {!selesai && mode === null && (
                <div className="flex gap-2">
                    <button type="button" disabled={busy} onClick={() => kirim("selesai")}
                        className="btn-primary h-8 px-3 text-xs disabled:opacity-50">Selesai</button>
                    <button type="button" disabled={busy} onClick={() => setMode("alasan")}
                        className="btn-secondary h-8 px-3 text-xs text-[#a66a17] disabled:opacity-50">
                        {o.konfirmasi ? "Ubah alasan" : "Belum terkirim"}
                    </button>
                </div>
            )}
            {mode === "alasan" && (
                <div className="flex flex-col gap-2">
                    <textarea className="input-field min-h-[64px] resize-none text-sm" value={alasan} maxLength={255}
                        onChange={(e) => setAlasan(e.target.value)} placeholder="Alasan (opsional), mis. pelanggan tidak di rumah" />
                    <div className="flex gap-2">
                        <button type="button" disabled={busy} onClick={() => kirim("belum_terkirim")}
                            className="btn-primary h-8 px-3 text-xs disabled:opacity-50">Simpan belum terkirim</button>
                        <button type="button" disabled={busy} onClick={() => setMode(null)}
                            className="btn-secondary h-8 px-3 text-xs">Batal</button>
                    </div>
                </div>
            )}
            {err && <p className="text-[#b74b40] text-xs">{err}</p>}
        </div>
    );
}

const KonfirmasiPengiriman = () => {
    const [view, setView] = useState(null);
    const [open, setOpen] = useState(false);

    const load = useCallback(async () => {
        try {
            const v = await api.getDeliveryConfirmations();
            setView(v);
            // buka otomatis sekali untuk tiap batch baru yang masih punya pesanan belum Selesai
            const pending = (v.batches || [])
                .filter((b) => b.orders.some((o) => o.konfirmasi?.status !== "selesai"))
                .map((b) => b.key);
            const key = `${v.date}:${pending.join(",")}`;
            if (pending.length > 0 && readSeen() !== key) {
                writeSeen(key);
                setOpen(true);
            }
        } catch {
            // diam saja: banner hanya muncul bila data berhasil dimuat
        }
    }, []);

    useEffect(() => {
        load();
        const t = setInterval(load, POLL_MS);
        return () => clearInterval(t);
    }, [load]);

    const batches = view?.batches || [];
    const pending = view?.pending || 0;
    // notifikasi selesai bila semua pesanan di batch yang dibuka sudah Selesai
    if (pending === 0 && !open) return null;

    return (
        <>
            {pending > 0 && (
                <button type="button" onClick={() => { setOpen(true); load(); }}
                    title="Pesanan di batch yang sudah dibuka dan belum Selesai"
                    className="flex items-center gap-2 h-9 px-3 rounded-lg text-xs font-semibold border transition-colors bg-[#fdf6ec] border-[#e8c98a] text-[#a66a17] hover:bg-[#fcebc8]">
                    Konfirmasi pengiriman
                    <span className="min-w-[18px] h-[18px] px-1 rounded-full bg-[#a66a17] text-white text-[10px] flex items-center justify-center">
                        {pending > 9 ? "9+" : pending}
                    </span>
                </button>
            )}

            {open && (
                <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
                    <div className="bg-white rounded-2xl shadow-modal w-[600px] flex flex-col max-h-[90vh]">
                        <div className="flex items-start justify-between px-6 py-5 border-b border-[#eadfd0]">
                            <div>
                                <h2 className="text-[#2b1d16] text-lg font-semibold">Konfirmasi pengiriman</h2>
                                <p className="text-[#8a7768] text-sm mt-0.5">
                                    Pastikan setiap pesanan sudah benar-benar sampai ke pelanggan.
                                </p>
                            </div>
                            <button onClick={() => setOpen(false)} className="text-[#8a7768] hover:text-[#2b1d16] text-xl leading-none">×</button>
                        </div>
                        <div className="overflow-y-auto flex flex-col gap-4 p-5">
                            {batches.length === 0 && (
                                <p className="text-[#166534] text-sm text-center py-6">Semua pesanan sudah dikonfirmasi Selesai.</p>
                            )}
                            {batches.map((b) => (
                                <div key={b.key} className="border border-[#eadfd0] rounded-xl overflow-hidden shrink-0">
                                    <div className="bg-[#faf6ee] px-5 py-2.5 flex items-center justify-between">
                                        <p className="text-[#2b1d16] text-sm font-semibold">{b.label}</p>
                                        <p className="text-[#7e4722] text-xs">{b.keterangan}</p>
                                    </div>
                                    {b.orders.map((o) => <OrderRow key={o.id} o={o} today={view.date} onDone={load} />)}
                                </div>
                            ))}
                        </div>
                        <div className="px-6 py-4 border-t border-[#eadfd0]">
                            <button onClick={() => setOpen(false)} className="btn-secondary w-full justify-center">Tutup</button>
                        </div>
                    </div>
                </div>
            )}
        </>
    );
};

export default KonfirmasiPengiriman;
