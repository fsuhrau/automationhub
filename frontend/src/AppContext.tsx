import React from "react";

export const AppContext = React.createContext({
    authenticated: false,
    activeMenu: "",
    lang: 'en',
})
