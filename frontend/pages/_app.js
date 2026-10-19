import React, { useEffect, useState } from "react";
import Head from "next/head";
import { Provider } from "react-redux";

import "../styles/globals.css";

import store from "../redux/store";
import { setUser } from "../redux/actions/userAction";

import Sidebar from "../components/Sidebar";
import Navbar from "../components/Navbar";
import { getSection } from "../components/Routes";

import { getToken, getNama } from "../utils/storage";

// halaman yang boleh dibuka tanpa login & tanpa sidebar/navbar
const PUBLIC_PAGES = ["/", "/login", "/404"];

const MyApp = ({ Component, pageProps, router }) => {
    const [mounted, setMounted] = useState(false);
    const isPublic = PUBLIC_PAGES.includes(router.pathname);

    useEffect(() => {
        setMounted(true);
        store.dispatch(setUser({ nama: getNama() }));
    }, []);

    // cek login setiap pindah halaman
    useEffect(() => {
        if (!isPublic && !getToken()) {
            router.replace("/login");
        }
    }, [router.pathname]);

    if (!mounted) return null;

    return (
        <>
            <Head>
                <meta name="viewport" content="width=device-width, initial-scale=1.0" />
                <title>Hakuu Bakery</title>
            </Head>

            <Provider store={store}>
                {isPublic ? (
                    <Component {...pageProps} />
                ) : !getToken() ? null : (
                    <div className="flex h-screen overflow-hidden bg-[#faf6ee] print:block print:h-auto print:overflow-visible print:bg-white">
                        <Sidebar />
                        <div className="flex flex-col flex-1 min-w-0 overflow-hidden print:overflow-visible">
                            <Navbar section={getSection(router.pathname)} />
                            <main className="flex-1 overflow-y-auto print:overflow-visible">
                                <Component {...pageProps} />
                            </main>
                        </div>
                    </div>
                )}
            </Provider>
        </>
    );
};

export default MyApp;
