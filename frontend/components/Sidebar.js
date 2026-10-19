import Link from "next/link";
import { useRouter } from "next/router";
import { useDispatch, useSelector } from "react-redux";

import { Routes } from "./Routes";
import { clearSession } from "../utils/storage";
import { clearUser } from "../redux/actions/userAction";
import { logoHakuuBakery, iconLogOut } from "../utils/icons";

const Sidebar = () => {
    const router = useRouter();
    const dispatch = useDispatch();
    const { user } = useSelector((state) => state.userReducer);

    const isActive = (path) => router.pathname === path || router.pathname.startsWith(path + "/");

    const handleLogout = () => {
        clearSession();
        dispatch(clearUser());
        router.push("/login");
    };

    return (
        <aside className="bg-[#4c0c13] flex flex-col gap-8 items-start p-4 h-screen w-60 shrink-0 sticky top-0 print:hidden">
            {/* Brand */}
            <div className="flex flex-col gap-4 items-start p-2 w-full">
                <img src={logoHakuuBakery} alt="Hakuu Bakery" className="h-[41px] w-[150px] object-contain" />
                <p className="text-[#fcebc8] text-xs">Penjadwalan Produksi Terpusat</p>
            </div>

            {/* Nav */}
            <nav className="flex flex-col gap-2 w-full flex-1">
                {Routes.map((item) => (
                    <Link
                        key={item.path}
                        href={item.path}
                        className={`nav-item ${isActive(item.path) ? "nav-item-active" : "nav-item-inactive"}`}
                    >
                        <img src={item.icon} alt="" className="size-4 shrink-0" />
                        <span className="text-sm whitespace-nowrap">{item.menu}</span>
                    </Link>
                ))}
            </nav>

            {/* User + logout — matches Figma: "Hakuu Bakery / Admin" */}
            <div className="border-t border-white/20 flex flex-col gap-4 items-start pt-4 w-full shrink-0">
                <div className="flex gap-3 items-center">
                    <div className="bg-[#daa278] flex items-center justify-center rounded-lg size-8 shrink-0">
                        <span className="text-[#4c0c13] text-xs font-semibold">HB</span>
                    </div>
                    <div className="flex flex-col">
                        <span className="text-white text-sm font-medium leading-5">{user.nama || "Hakuu Bakery"}</span>
                        <span className="text-[#fcebc8] text-xs leading-4">Admin</span>
                    </div>
                </div>
                <button
                    onClick={handleLogout}
                    className="border border-white/30 flex gap-3 h-10 items-center px-4 rounded-lg w-full text-white text-sm hover:bg-white/10 transition-colors"
                >
                    <img src={iconLogOut} alt="" className="size-4 shrink-0" />
                    <span>Keluar</span>
                </button>
            </div>
        </aside>
    );
};

export default Sidebar;
