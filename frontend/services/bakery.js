import apiClient from "../utils/apiClient";

const enc = encodeURIComponent;

// =============================================================================================
// Pesanan (Orders)
// =============================================================================================
const getOrders = async (limit = 200) => {
    return await apiClient.get(`/api/orders?limit=${limit}`);
};

const getOrder = async (id) => {
    return await apiClient.get(`/api/orders/${id}`);
};

const insertOrder = async (body) => {
    return await apiClient.post("/api/orders", body);
};

// action: start | complete | cancel
const orderAction = async (id, action) => {
    return await apiClient.post(`/api/orders/${id}/${action}`);
};

// =============================================================================================
// Penawaran (Offers / Negosiasi)
// =============================================================================================
const offerToBuyer = async (offerId) => {
    return await apiClient.post(`/api/offers/${offerId}/offer`);
};

const respondOffer = async (offerId, accept) => {
    return await apiClient.post(`/api/offers/${offerId}/respond`, { accept });
};

// =============================================================================================
// Jadwal Produksi
// =============================================================================================
const getSchedule = async (from, days = 7) => {
    return await apiClient.get(`/api/schedule?from=${from}&days=${days}`);
};

// =============================================================================================
// Pengaturan
// =============================================================================================
const getSettings = async () => {
    return await apiClient.get("/api/settings");
};

const updateSettings = async (body) => {
    return await apiClient.put("/api/settings", body);
};

// =============================================================================================
// Produk
// =============================================================================================
const getProducts = async () => {
    return await apiClient.get("/api/products");
};

const insertProduct = async (body) => {
    return await apiClient.post("/api/products", body);
};

const updateProduct = async (id, body) => {
    return await apiClient.patch(`/api/products/${id}`, body);
};

const deleteProduct = async (id) => {
    return await apiClient.delete(`/api/products/${id}`);
};

// =============================================================================================
// Kategori
// =============================================================================================
const getCategories = async () => {
    return await apiClient.get("/api/categories");
};

const insertCategory = async (nama) => {
    return await apiClient.post("/api/categories", { nama });
};

const deleteCategory = async (nama) => {
    return await apiClient.delete(`/api/categories/${enc(nama)}`);
};

// =============================================================================================
// Notifikasi
// =============================================================================================
// respons: [{ id, pesan, waktu }] (hanya yang belum dibaca)
const getNotifications = async () => {
    return await apiClient.get("/api/notifications");
};

const readNotifications = async () => {
    return await apiClient.post("/api/notifications/read");
};

// =============================================================================================
// Laporan Bulanan
// =============================================================================================
// month: 'YYYY-MM'
const getMonthlyReport = async (month) => {
    return await apiClient.get(`/api/reports/monthly?month=${enc(month)}`);
};

// file CSV (Blob)
const exportMonthlyReport = async (month) => {
    return await apiClient.get(`/api/reports/monthly/export?month=${enc(month)}`, { responseType: "blob" });
};

const closeMonth = async (month) => {
    return await apiClient.post(`/api/reports/monthly/${enc(month)}/close`);
};

export default {
    getOrders,
    getOrder,
    insertOrder,
    orderAction,
    offerToBuyer,
    respondOffer,
    getSchedule,
    getSettings,
    updateSettings,
    getProducts,
    insertProduct,
    updateProduct,
    deleteProduct,
    getCategories,
    insertCategory,
    deleteCategory,
    getNotifications,
    readNotifications,
    getMonthlyReport,
    exportMonthlyReport,
    closeMonth,
};
