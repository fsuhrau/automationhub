import { createContext, ReactElement, useState } from 'react';
import ITestData from '../types/test';

type TestContextProps = {
    test: ITestData,
    setTest: (t: ITestData) => {},
};

export const TestContext = createContext<Partial<TestContextProps>>({});
/*
export const TestContextProvider = (props: TestContextProps): ReactElement => {
    const { children } = props;
    const [ test, setTest ] = useState<ITestData>();
    return (
        <TestContext.Provider value={ { test, setTest }}>
            {children}
        </TestContext.Provider>
    );
};
*/