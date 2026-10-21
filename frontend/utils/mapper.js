/*
   note : pemetaan data dari backend -> bentuk yang dipakai di tampilan (UI)
*/
import { uiStatus, ymd } from "./orderStatus";

export const JENIS_LABEL = { tanpa_topping: "Tanpa topping", bertopping: "Bertopping", campuran: "Campuran" };
export const JENIS_VALUE = { "Tanpa topping": "tanpa_topping", Bertopping: "bertopping", Campuran: "campuran" };

/**
 * Bobot kapasitas per box (sama dengan HitungBobot di backend).
 * Patokan: box isi 3 bertopping = 1,23, jadi 1 roti bertopping = 0,41.
 * Tanpa topping 1,00 per box; bertopping 0,41 x isi; campuran 0,41 x roti bertopping + roti tanpa topping / isi.
 */
export const BOBOT_PER_ROTI_TOPPING = 1.23 / 3;
export function hitungBobot(jenis, isi, topping) {
    let b = 1;
    if (jenis === "Bertopping") b = BOBOT_PER_ROTI_TOPPING * isi;
    if (jenis === "Campuran") b = BOBOT_PER_ROTI_TOPPING * topping + (isi - topping) / isi;
    return Math.round(b * 100) / 100;
}

/** Pesan galat isi box (string kosong = valid). */
export function komposisiProblem(jenis, isi, topping) {
    if (!Number.isInteger(isi) || isi < 1 || isi > 50) return "Isi box harus 1 sampai 50 roti.";
    if (jenis === "Campuran" && (!Number.isInteger(topping) || topping < 1 || topping >= isi))
        return `Produk campuran berisi roti bertopping dan tanpa topping, jadi roti bertopping harus 1 sampai ${isi - 1}.`;
    if (hitungBobot(jenis, isi, topping) > 5) return "Bobot maksimal 5, kurangi isi box.";
    return "";
}

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
        isiBox: p.isi_box || 3,
        isiTopping: p.isi_topping || 0,
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
