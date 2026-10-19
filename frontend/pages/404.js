import { useEffect } from "react";
import { useRouter } from "next/router";

// halaman tidak ditemukan -> kembali ke "/"
const NotFound = () => {
    const router = useRouter();

    useEffect(() => {
        router.replace("/");
    }, []);

    return null;
};

export default NotFound;
