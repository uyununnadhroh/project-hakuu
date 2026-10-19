/**
 * Pilih varian untuk satu paket.
 *   variants : [{ id, name, optionLabel, options }]
 *   max      : jumlah varian yang wajib dipilih (= isi paket)
 *   selected : [{ id, opsi }]
 * Saat batas tercapai, chip yang belum dipilih nonaktif; yang sudah dipilih tetap bisa dibatalkan.
 */
function IconCheck() {
    return (
        <svg viewBox="0 0 24 24" className="size-3.5 shrink-0" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
            <path d="M20 6 9 17l-5-5" />
        </svg>
    );
}

const VariantPicker = ({ variants, max, selected, onChange }) => {
    const isSelected = (id) => selected.some((s) => s.id === id);
    const full = selected.length >= max;
    const sisa = max - selected.length;
    const opsiKosong = selected
        .map((s) => variants.find((v) => v.id === s.id))
        .filter((v, i) => v && v.options.length > 0 && !selected[i].opsi);

    function toggle(v) {
        if (isSelected(v.id)) onChange(selected.filter((s) => s.id !== v.id));
        else if (!full) onChange([...selected, { id: v.id, opsi: "" }]);
    }

    function setOpsi(id, opsi) {
        onChange(selected.map((s) => (s.id === id ? { ...s, opsi } : s)));
    }

    let helper;
    if (sisa > 0) helper = `Pilih ${sisa} varian lagi`;
    else if (opsiKosong.length > 0) helper = `Lengkapi ${opsiKosong[0].optionLabel || "pilihan"} untuk ${opsiKosong[0].name}`;
    else helper = "Varian lengkap";
    const lengkap = sisa === 0 && opsiKosong.length === 0;

    return (
        <div className="flex flex-col gap-3 border-t border-[#eadfd0] pt-4">
            <div className="flex items-center justify-between">
                <p className="text-[#2b1d16] text-sm font-semibold">Pilih varian</p>
                <span className="text-[#7e4722] text-xs font-medium">{selected.length}/{max} dipilih</span>
            </div>

            {variants.length === 0 ? (
                <p className="text-[#b74b40] text-xs">Paket ini belum punya varian. Tambahkan varian di database terlebih dahulu.</p>
            ) : (
                <div className="flex gap-2 flex-wrap">
                    {variants.map((v) => {
                        const on = isSelected(v.id);
                        const disabled = !on && full;
                        return (
                            <button
                                key={v.id}
                                type="button"
                                onClick={() => toggle(v)}
                                disabled={disabled}
                                aria-pressed={on}
                                className={`inline-flex items-center gap-1.5 px-3.5 h-8 rounded-full text-xs font-semibold transition-colors ${
                                    on
                                        ? "bg-[#4c0c13] text-white border border-[#4c0c13]"
                                        : "bg-white text-[#2b1d16] border border-[#eadfd0] hover:bg-[#faf6ee]"
                                } disabled:opacity-40 disabled:cursor-not-allowed disabled:hover:bg-white`}
                            >
                                {on && <IconCheck />}
                                {v.name}
                            </button>
                        );
                    })}
                </div>
            )}

            {/* opsi tambahan, hanya untuk varian yang dipilih */}
            {selected.map((s) => {
                const v = variants.find((x) => x.id === s.id);
                if (!v || v.options.length === 0) return null;
                return (
                    <div key={v.id} className="flex items-center gap-3">
                        <label className="text-[#2b1d16] text-sm w-[140px] shrink-0">
                            {v.optionLabel || "Pilihan"} ({v.name})
                        </label>
                        <select className="input-field" value={s.opsi} onChange={(e) => setOpsi(v.id, e.target.value)}>
                            <option value="">Pilih {(v.optionLabel || "pilihan").toLowerCase()}…</option>
                            {v.options.map((o) => <option key={o} value={o}>{o}</option>)}
                        </select>
                    </div>
                );
            })}

            <p className={`text-xs ${lengkap ? "text-[#327a52]" : "text-[#a66a17]"}`}>{helper}</p>
        </div>
    );
};

export default VariantPicker;
