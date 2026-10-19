import { useCallback, useEffect, useState } from "react";

/** Muat data async; `reload()` mengulang. Pesan galat sudah berbahasa Indonesia. */
export function useLoad(fn, deps = []) {
    const [data, setData] = useState(null);
    const [error, setError] = useState("");
    const [loading, setLoading] = useState(true);

    // eslint-disable-next-line react-hooks/exhaustive-deps
    const run = useCallback(async () => {
        setLoading(true);
        setError("");
        try {
            setData(await fn());
        } catch (e) {
            setError(e.message || "Gagal memuat data");
        } finally {
            setLoading(false);
        }
    }, deps);

    useEffect(() => { run(); }, [run]);
    return { data, error, loading, reload: run };
}

export default useLoad;
