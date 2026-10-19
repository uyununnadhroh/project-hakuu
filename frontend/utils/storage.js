/*
   note : - simpan / ambil sesi login (token, nama & email user)
          - remember=true  -> localStorage (tetap login setelah browser ditutup)
            remember=false -> sessionStorage
*/

const TOKEN_KEY = "hakuu_token";
const NAME_KEY = "hakuu_nama";
const EMAIL_KEY = "hakuu_email";

const isBrowser = () => typeof window !== "undefined";

export const getToken = () =>
    isBrowser() ? localStorage.getItem(TOKEN_KEY) || sessionStorage.getItem(TOKEN_KEY) : null;

export const getNama = () =>
    isBrowser() ? localStorage.getItem(NAME_KEY) || sessionStorage.getItem(NAME_KEY) || "" : "";

export const getEmail = () =>
    isBrowser() ? localStorage.getItem(EMAIL_KEY) || sessionStorage.getItem(EMAIL_KEY) || "" : "";

export function setSession(token, nama, remember, email = "") {
    clearSession();
    const store = remember ? localStorage : sessionStorage;
    store.setItem(TOKEN_KEY, token);
    store.setItem(NAME_KEY, nama || "");
    store.setItem(EMAIL_KEY, email || "");
}

export function clearSession() {
    if (!isBrowser()) return;
    for (const s of [localStorage, sessionStorage]) {
        s.removeItem(TOKEN_KEY);
        s.removeItem(NAME_KEY);
        s.removeItem(EMAIL_KEY);
    }
}
