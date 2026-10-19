import { useCallback, useEffect, useRef, useState } from "react";
import { iconStore, iconBell } from "../utils/icons";
import api from "../services/bakery";
import KonfirmasiPengiriman from "./KonfirmasiPengiriman";

// notifikasi diperbarui otomatis tiap 60 detik
const POLL_MS = 60000;

const Navbar = ({ section = "Dashboard Admin" }) => {
    const today = new Date().toLocaleDateString("id-ID", {
        weekday: "long",
        day: "numeric",
        month: "long",
        year: "numeric",
    });

    const [notifs, setNotifs] = useState([]);
    const [open, setOpen] = useState(false);
    const [error, setError] = useState("");
    const [busy, setBusy] = useState(false);
    const boxRef = useRef(null);

    const load = useCallback(async () => {
        try {
            const r = await api.getNotifications();
            setNotifs(Array.isArray(r) ? r : []);
            setError("");
        } catch (e) {
            setError(e.message);
        }
    }, []);

    useEffect(() => {
        load();
        const t = setInterval(load, POLL_MS);
        return () => clearInterval(t);
    }, [load]);

    // tutup panel bila klik di luar
    useEffect(() => {
        if (!open) return;
        const onClick = (e) => { if (boxRef.current && !boxRef.current.contains(e.target)) setOpen(false); };
        document.addEventListener("mousedown", onClick);
        return () => document.removeEventListener("mousedown", onClick);
    }, [open]);

    async function handleReadAll() {
        setBusy(true);
        try {
            await api.readNotifications();
            setNotifs([]);
        } catch (e) {
            setError(e.message);
        } finally {
            setBusy(false);
        }
    }

    return (
        <header className="bg-white border-b border-[#eadfd0] flex h-20 items-center justify-between px-9 w-full shrink-0 print:hidden">
            <div className="flex gap-2.5 items-center">
                <img src={iconStore} alt="" className="size-5" />
                <span className="text-[#2b1d16] text-sm font-semibold">Hakuu Bakery</span>
                <span className="text-[#8a7768] text-sm">/ {section}</span>
            </div>
            <div className="flex gap-4 items-center">
                <span className="text-[#8a7768] text-xs">{today}</span>
                <KonfirmasiPengiriman />
                <div className="relative" ref={boxRef}>
                    <button
                        type="button"
                        className="relative flex items-center justify-center size-9 rounded-lg hover:bg-[#faf6ee]"
                        onClick={() => { setOpen(!open); if (!open) load(); }}
                        aria-label={`Notifikasi${notifs.length ? ` (${notifs.length} belum dibaca)` : ""}`}
                    >
                        <img src={iconBell} alt="" className="size-5" />
                        {notifs.length > 0 && (
                            <span className="absolute -top-0.5 -right-0.5 min-w-[18px] h-[18px] px-1 rounded-full bg-[#b74b40] text-white text-[10px] font-semibold flex items-center justify-center">
                                {notifs.length > 9 ? "9+" : notifs.length}
                            </span>
                        )}
                    </button>

                    {open && (
                        <div className="absolute right-0 top-11 z-50 w-[360px] bg-white border border-[#eadfd0] rounded-xl shadow-modal flex flex-col max-h-[420px]">
                            <div className="flex items-center justify-between px-4 py-3 border-b border-[#eadfd0]">
                                <p className="text-[#2b1d16] text-sm font-semibold">Notifikasi</p>
                                {notifs.length > 0 && (
                                    <button type="button" disabled={busy} onClick={handleReadAll}
                                        className="text-[#4c0c13] text-xs font-semibold hover:underline disabled:opacity-50">
                                        Tandai semua dibaca
                                    </button>
                                )}
                            </div>
                            <div className="overflow-y-auto">
                                {error && <p className="px-4 py-3 text-xs text-[#b74b40]">{error}</p>}
                                {!error && notifs.length === 0 && (
                                    <p className="px-4 py-6 text-center text-[#8a7768] text-sm">Tidak ada notifikasi baru.</p>
                                )}
                                {notifs.map((n) => (
                                    <div key={n.id} className="px-4 py-3 border-t first:border-t-0 border-[#eadfd0]">
                                        <p className="text-[#2b1d16] text-sm">{n.pesan}</p>
                                        <p className="text-[#8a7768] text-xs mt-0.5">{n.waktu}</p>
                                    </div>
                                ))}
                            </div>
                        </div>
                    )}
                </div>
                <div className="bg-[#faf6ee] flex items-center justify-center rounded-full size-9">
                    <span className="text-[#4c0c13] text-sm font-semibold">HB</span>
                </div>
                <span className="text-[#2b1d16] text-sm font-semibold">Hakuu Bakery</span>
            </div>
        </header>
    );
};

export default Navbar;
