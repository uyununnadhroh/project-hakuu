/*
   note : - instance axios yang dipakai semua file di folder services/
          - token otomatis dipasang di header Authorization
          - response langsung dikembalikan isi body-nya (response.data)
          - pesan error diseragamkan (bahasa Indonesia)
*/

import axios from "axios";
import { getToken, clearSession } from "./storage";

export class ApiError extends Error {
    constructor(message, status, data) {
        super(message);
        this.status = status;
        this.data = data;
    }
}

const apiClient = axios.create({
    baseURL: (process.env.NEXT_PUBLIC_BASE_API_URL || "http://localhost:8080").replace(/\/$/, ""),
    timeout: 60000,
});

apiClient.interceptors.request.use(
    (config) => {
        const token = getToken();
        if (token) config.headers.Authorization = `Bearer ${token}`;
        return config;
    },
    (error) => Promise.reject(error)
);

apiClient.interceptors.response.use(
    (response) => response.data,
    (error) => {
        // backend tidak bisa dihubungi
        if (!error.response) {
            return Promise.reject(
                new ApiError("Tidak dapat terhubung ke server. Pastikan backend sedang berjalan.", 0, null)
            );
        }

        const { status, data } = error.response;

        // sesi habis -> balik ke login
        if (status === 401 && error.config?.url !== "/api/login") {
            clearSession();
            window.location.assign("/login");
            return Promise.reject(new ApiError("Sesi berakhir, silakan masuk lagi.", 401, data));
        }

        return Promise.reject(
            new ApiError(data?.error || data?.reason || `Permintaan gagal (${status})`, status, data)
        );
    }
);

export default apiClient;
