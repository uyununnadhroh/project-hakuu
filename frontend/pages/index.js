import { useEffect } from "react";
import { useRouter } from "next/router";
import { getToken } from "../utils/storage";

// "/" -> ke dashboard kalau sudah login, kalau belum ke halaman login
const Home = () => {
    const router = useRouter();

    useEffect(() => {
        router.replace(getToken() ? "/dashboard" : "/login");
    }, []);

    return null;
};

export default Home;
