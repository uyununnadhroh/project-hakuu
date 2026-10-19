import apiClient from "../utils/apiClient";

// =============================================================================================
// Auth
// =============================================================================================
const login = async (email, password) => {
    return await apiClient.post("/api/login", { email, password });
};

// daftar akun (email) yang boleh masuk ke website
// respons yang diharapkan: [{ id, nama, email, created_at }]
const getUsers = async () => {
    return await apiClient.get("/api/users");
};

export default {
    login,
    getUsers,
};
