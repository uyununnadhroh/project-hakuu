import QtyStepper from "./QtyStepper";
import VariantPicker from "./VariantPicker";

/**
 * Kartu satu paket pilih varian.
 *   pkg  : produk hasil toUiProduct (maxVariants > 0)
 *   item : baris di orderItems untuk paket ini (undefined bila qty 0)
 * Qty 0 -> bagian varian disembunyikan dan pilihannya di-reset (baris dihapus dari orderItems).
 */
const PackageCard = ({ pkg, item, onQty, onVariants }) => {
    const qty = item?.qty ?? 0;
    const aktif = qty > 0;

    return (
        <div className={`card p-5 flex flex-col gap-4 transition-colors ${aktif ? "border-[#4c0c13]" : ""}`}>
            <div className="flex items-center justify-between gap-4">
                <div className="flex-1 min-w-0">
                    <p className="text-[#2b1d16] text-sm font-semibold">{pkg.name}</p>
                    <p className="text-[#8a7768] text-xs mt-0.5">
                        Isi {pkg.maxVariants} varian · Rp {pkg.price.toLocaleString("id-ID")} / box
                    </p>
                </div>
                <QtyStepper value={qty} onChange={onQty} label={`Jumlah box ${pkg.name}`} />
            </div>

            {aktif && (
                <VariantPicker
                    variants={pkg.variants}
                    max={pkg.maxVariants}
                    selected={item.chosenVariants || []}
                    onChange={onVariants}
                />
            )}
        </div>
    );
};

export default PackageCard;
