import { createContext } from 'react';

type AppContextProps = {
};

export const AppContext = createContext<Partial<AppContextProps>>({ });
