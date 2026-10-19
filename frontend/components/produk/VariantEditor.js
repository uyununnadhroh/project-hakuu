/**
 * Atur isi paket & daftar varian (Kelola Produk).
 *   maxVariants : isi paket (jumlah varian yang dipilih pembeli per box)
 *   variants    : [{ id, name, optionLabel, optionsText, active }]
 *                 optionsText = pilihan tambahan dipisah koma, mis. "Stroberi, Blueberry"
 * Varian yang dihapus tapi sudah pernah dipesan akan dinonaktifkan oleh server.
 */
export const emptyVariant = () => ({ id: undefined, name: "", optionLabel: "", optionsText: "", active: true, showOpsi: false });

/** Data varian dari toUiProduct -> baris editor. */
export const toEditorVariant = (v) => ({
    id: v.id,
    name: v.name,
    optionLabel: v.optionLabel,
    optionsText: v.options.join(", "),
    active: v.active,
    showOpsi: v.options.length > 0,
});

/** Baris editor -> payload API. */
export const toVariantPayload = (v) => ({
    id: v.id,
    nama: v.name.trim(),
    opsi_label: v.optionLabel.trim(),
    opsi_pilihan: v.optionsText.split(",").map((s) => s.trim()).filter(Boolean),
    aktif: v.active,
});

/** Pesan kenapa daftar varian belum bisa disimpan, atau "" bila sudah benar. */
export function variantProblem(maxVariants, variants) {
    const maks = Number(maxVariants);
    if (!Number.isInteger(maks) || maks < 1 || maks > 10) return "Isi paket harus 1 sampai 10 varian.";
    if (variants.some((v) => !v.name.trim())) return "Nama varian wajib diisi.";
    const names = variants.map((v) => v.name.trim().toLowerCase());
    if (new Set(names).size !== names.length) return "Ada nama varian yang sama.";
    const aktif = variants.filter((v) => v.active).length;
    if (aktif < maks) return `Butuh minimal ${maks} varian aktif (baru ${aktif}).`;
    return "";
}

const VariantEditor = ({ maxVariants, variants, onMaxChange, onChange }) => {
    const set = (i, patch) => onChange(variants.map((v, j) => (j === i ? { ...v, ...patch } : v)));
    const remove = (i) => onChange(variants.filter((_, j) => j !== i));
    const problem = variantProblem(maxVariants, variants);

    return (
        <div className="flex flex-col gap-4 border border-[#eadfd0] rounded-xl p-4 bg-[#faf6ee]/50">
            <div className="flex flex-col gap-2">
                <label className="text-[#2b1d16] text-sm">Isi paket (varian per box)</label>
                <input className="input-field w-[120px]" type="number" min={1} max={10} value={maxVariants}
                    onChange={(e) => onMaxChange(e.target.value)} placeholder="3" />
                <p className="text-[#8a7768] text-xs">Pembeli wajib memilih tepat sejumlah ini saat input pesanan.</p>
            </div>

            <div className="flex flex-col gap-2">
                <p className="text-[#2b1d16] text-sm">Daftar varian</p>
                {variants.length === 0 && <p className="text-[#8a7768] text-xs">Belum ada varian.</p>}
                {variants.map((v, i) => (
                    <div key={v.id ?? `baru-${i}`} className="bg-white border border-[#eadfd0] rounded-xl p-3 flex flex-col gap-2">
                        <div className="flex items-center gap-2">
                            <input className="input-field flex-1" value={v.name} placeholder="Mis. Coklat"
                                onChange={(e) => set(i, { name: e.target.value })} aria-label={`Nama varian ${i + 1}`} />
                            <button type="button" onClick={() => set(i, { active: !v.active })}
                                className={`h-10 px-3 rounded-lg border text-xs font-semibold shrink-0 ${
                                    v.active ? "border-[#4c0c13] text-[#4c0c13] bg-[#4c0c13]/5" : "border-[#eadfd0] text-[#8a7768]"
                                }`}>
                                {v.active ? "Aktif" : "Nonaktif"}
                            </button>
                            <button type="button" onClick={() => remove(i)}
                                className="text-[#dc2626] text-xs hover:underline shrink-0 px-1">Hapus</button>
                        </div>
                        {v.showOpsi ? (
                            <div className="flex gap-2">
                                <input className="input-field w-[150px]" value={v.optionLabel} placeholder="Isian selai"
                                    onChange={(e) => set(i, { optionLabel: e.target.value })} aria-label="Label pilihan tambahan" />
                                <input className="input-field flex-1" value={v.optionsText} placeholder="Stroberi, Blueberry, Nanas"
                                    onChange={(e) => set(i, { optionsText: e.target.value })} aria-label="Pilihan tambahan, pisahkan dengan koma" />
                            </div>
                        ) : (
                            <button type="button" onClick={() => set(i, { showOpsi: true })}
                                className="text-[#4c0c13] text-xs font-semibold hover:underline self-start">
                                + Pilihan tambahan (mis. isian selai)
                            </button>
                        )}
                    </div>
                ))}
                <button type="button" onClick={() => onChange([...variants, emptyVariant()])}
                    className="btn-secondary text-[#4c0c13] self-start h-9 text-sm">
                    + Tambah varian
                </button>
            </div>

            {problem && <p className="text-[#a66a17] text-xs">{problem}</p>}
        </div>
    );
};

export default VariantEditor;
