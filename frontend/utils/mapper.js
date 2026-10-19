/*
   note : pemetaan data dari backend -> bentuk yang dipakai di tampilan (UI)
*/
import { uiStatus, ymd } from "./orderStatus";

export const JENIS_LABEL = { tanpa_topping: "Tanpa topping", bertopping: "Bertopping", campuran: "Campuran" };
export const JENIS_VALUE = { "Tanpa topping": "tanpa_topping", Bertopping: "bertopping", Campuran: "campuran" };

/**
 * Penanggung jawab = akun yang menginput pesanan (diisi backend dari token login).
 * Field yang diharapkan dari backend: created_by_nama & created_by_email.
 * Beberapa nama field lain ikut dibaca supaya tetap jalan bila backend memakai penamaan berbeda.
 */
export function penanggungJawab(o = {}) {
    const obj = typeof o.created_by === "object" && o.created_by ? o.created_by : {};
    const nama = o.created_by_nama || o.created_by_name || o.penanggung_jawab || obj.nama || obj.name || "";
    const email = o.created_by_email || o.penanggung_jawab_email || obj.email || "";
    return { nama, email, label: nama || email || "" };
}

export function toUiOrder(o, today = ymd()) {
    const pj = penanggungJawab(o);
    return {
        dbId: o.id,
        id: o.kode,
        pembeli: o.customer || "—",
        produk: o.products || "—",
        qty: o.qty_box,            // jumlah box dipesan
        beban: o.qty,              // beban kapasitas (box x bobot)
        tanggalKirim: o.deliver_at.slice(0, 10),
        jamKirim: o.deliver_at.slice(11, 16),
        tanggalProduksi: o.production_day,
        waktuMasuk: o.ordered_at,
        skor: o.score,
        backendStatus: o.status,
        canStart: o.can_start,
        needsNego: o.needs_negotiation,
        statusOtomatis: uiStatus(o, today),
        pj: pj.label,              // nama (atau email) penanggung jawab
        pjEmail: pj.email,
    };
}

export function toUiProduct(p) {
    return {
        id: p.id,
        name: p.nama,
        price: p.harga,
        active: p.aktif,
        kategori: p.kategori,
        jenis: JENIS_LABEL[p.jenis] || p.jenis,
        bobot: p.bobot,
        // paket pilih varian: maxVariants > 0 = wajib memilih tepat sejumlah ini
        maxVariants: p.maks_varian || 0,
        variants: (p.variants || []).map((v) => ({
            id: v.id,
            name: v.nama,
            optionLabel: v.opsi_label || "",
            options: v.opsi_pilihan || [],
            active: v.aktif !== false,
        })),
    };
}

/** Pilihan varian satu paket sudah lengkap? (jumlah pas + opsi tambahan terisi) */
export function variantsComplete(item) {
    if (!item.maxVariants) return true;
    const chosen = item.chosenVariants || [];
    if (chosen.length !== item.maxVariants) return false;
    return chosen.every((c) => {
        const v = item.variants.find((x) => x.id === c.id);
        return !v || v.options.length === 0 || !!c.opsi;
    });
}

/** Teks varian terpilih untuk ringkasan, mis. "Coklat, Lava (Stroberi)" */
export function variantSummary(item) {
    return (item.chosenVariants || [])
        .map((c) => {
            const v = item.variants.find((x) => x.id === c.id);
            const name = v ? v.name : `#${c.id}`;
            return c.opsi ? `${name} (${c.opsi})` : name;
        })
        .join(", ");
}
