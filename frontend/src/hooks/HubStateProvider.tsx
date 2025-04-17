import React, {createContext, useContext, useReducer, Dispatch, ReactNode} from 'react';
import { hubReducer, HubState, InitialHubState, HubStateAction } from '../application/HubState';

interface HubStateProviderProps {
    children: ReactNode;
}

interface HubStateContextProps {
    state: HubState;
    dispatch: Dispatch<HubStateAction>;
}

const HubStateContext = createContext<HubStateContextProps>({
    state: InitialHubState,
    dispatch: () => undefined,
});

export const HubStateProvider: React.FC<HubStateProviderProps> = ({ children }) => {
    const [state, dispatch] = useReducer(hubReducer, InitialHubState);

    return (
        <HubStateContext.Provider value={{ state, dispatch }}>
            {children}
        </HubStateContext.Provider>
    );
};

export const useHubState = (): HubStateContextProps => {
    return useContext(HubStateContext);
};