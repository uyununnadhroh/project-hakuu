import { useState } from "react";

/**
 * Stepper jumlah box: −, angka (bisa diketik), +.
 * Tidak bisa negatif. Kotak angka boleh kosong sementara diketik; saat ditinggalkan kembali ke `min`.
 */
const QtyStepper = ({ value, onChange, min = 0, label = "Jumlah box" }) => {
    const [draft, setDraft] = useState(null); // teks yang sedang diketik

    const set = (n) => onChange(Math.max(min, n));

    function handleTyping(raw) {
        const text = raw.replace(/[^0-9]/g, "");
        setDraft(text);
        const n = parseInt(text, 10);
        if (!isNaN(n)) set(n);
    }

    function commit() {
        if (draft !== null && isNaN(parseInt(draft, 10))) set(min);
        setDraft(null);
    }

    return (
        <div className="flex items-center gap-1.5">
            <button
                type="button"
                onClick={() => { setDraft(null); set(value - 1); }}
                disabled={value <= min}
                aria-label={`Kurangi ${label}`}
                className="size-7 rounded-full border border-[#eadfd0] text-[#4c0c13] font-bold text-lg leading-none hover:bg-[#faf6ee] shrink-0 disabled:opacity-40 disabled:cursor-not-allowed"
            >
                −
            </button>
            <input
                type="text"
                inputMode="numeric"
                aria-label={label}
                value={draft ?? String(value)}
                onFocus={(e) => e.target.select()}
                onChange={(e) => handleTyping(e.target.value)}
                onBlur={commit}
                onKeyDown={(e) => { if (e.key === "Enter") e.currentTarget.blur(); }}
                className="w-12 text-center text-sm font-semibold text-[#2b1d16] border border-[#eadfd0] rounded-lg py-1 focus:outline-none focus:border-[#4c0c13] [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
            />
            <button
                type="button"
                onClick={() => { setDraft(null); set(value + 1); }}
                aria-label={`Tambah ${label}`}
                className="size-7 rounded-full bg-[#4c0c13] text-white font-bold text-lg leading-none hover:bg-[#3a0a0f] shrink-0"
            >
                +
            </button>
        </div>
    );
};

export default QtyStepper;
