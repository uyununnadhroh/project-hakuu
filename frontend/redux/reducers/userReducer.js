const initialState = {
    user: {
        nama: "",
    },
};

const userReducer = (state = initialState, action) => {
    switch (action.type) {
        case "SET_USER":
            return { ...state, user: { ...state.user, ...action.payload } };
        case "CLEAR_USER":
            return initialState;
        default:
            return state;
    }
};

export default userReducer;
