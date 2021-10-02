import { createContext } from 'react';

type ContextProps = {
};

export const AppContext = createContext<Partial<ContextProps>>({ });
