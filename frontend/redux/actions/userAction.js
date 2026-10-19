export const setUser = (payload) => (dispatch) => {
    return dispatch({ type: "SET_USER", payload });
};

export const clearUser = () => (dispatch) => {
    return dispatch({ type: "CLEAR_USER" });
};
