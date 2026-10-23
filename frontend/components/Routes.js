/*
   note : daftar menu sidebar + judul yang tampil di Navbar (section)
          tambah menu baru cukup di sini
*/
import {
    iconLayoutDashboard,
    iconClipboardList,
    iconCalendarDays,
    iconCalendarCheck,
    iconPackage,
    iconSettings,
    iconLaporan,
} from "../utils/icons";

export const Routes = [
    { path: "/dashboard",       menu: "Dashboard",       section: "Dashboard Admin", icon: iconLayoutDashboard },
    { path: "/input-pesanan",   menu: "Input Pesanan",   section: "Input Pesanan",   icon: iconClipboardList   },
    { path: "/daftar-pesanan",  menu: "Daftar Pesanan",  section: "Daftar Pesanan",  icon: iconCalendarDays    },
    { path: "/jadwal-produksi", menu: "Jadwal Produksi", section: "Jadwal Produksi", icon: iconCalendarCheck   },
    { path: "/laporan-penjualan", menu: "Laporan Penjualan", section: "Laporan Penjualan", icon: iconLaporan },
    { path: "/kelola-produk",   menu: "Kelola Produk",   section: "Kelola Produk",   icon: iconPackage         },
    { path: "/pengaturan",      menu: "Pengaturan",      section: "Pengaturan",      icon: iconSettings        },
];

export const getSection = (pathname) =>
    Routes.find((r) => r.path === pathname)?.section ?? "Dashboard Admin";
