import { iconStore, iconBell } from "../utils/icons";

const Navbar = ({ section = "Dashboard Admin" }) => {
    const today = new Date().toLocaleDateString("id-ID", {
        weekday: "long",
        day: "numeric",
        month: "long",
        year: "numeric",
    });

    return (
        <header className="bg-white border-b border-[#eadfd0] flex h-20 items-center justify-between px-9 w-full shrink-0 print:hidden">
            <div className="flex gap-2.5 items-center">
                <img src={iconStore} alt="" className="size-5" />
                <span className="text-[#2b1d16] text-sm font-semibold">Hakuu Bakery</span>
                <span className="text-[#8a7768] text-sm">/ {section}</span>
            </div>
            <div className="flex gap-4 items-center">
                <span className="text-[#8a7768] text-xs">{today}</span>
                <button className="relative">
                    <img src={iconBell} alt="Notifikasi" className="size-5" />
                </button>
                <div className="bg-[#faf6ee] flex items-center justify-center rounded-full size-9">
                    <span className="text-[#4c0c13] text-sm font-semibold">HB</span>
                </div>
                <span className="text-[#2b1d16] text-sm font-semibold">Hakuu Bakery</span>
            </div>
        </header>
    );
};

export default Navbar;
