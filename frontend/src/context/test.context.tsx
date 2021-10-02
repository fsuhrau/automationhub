import { createContext, useState } from 'react';
import ITestData from "../types/test";

type ContextProps = {
    test: ITestData,
    setTest: (t: ITestData) => {},
};

export const TestContext = createContext<ContextProps>({});

export const TestContextProvider = (props) => {
    const { children } = props;
    const [ test, setTest ] = useState<ITestData>();
    return (
        <TestContext.Provider value={ { test, setTest }}>
            {children}
        </TestContext.Provider>
    )
};
