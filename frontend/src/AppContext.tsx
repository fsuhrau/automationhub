import { createContext } from 'react';

export const AppContext = createContext({
    authenticated: false,
    activeMenu: '',
    lang: 'en',
});
