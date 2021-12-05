import { createContext, ReactElement, useState } from 'react';
import ITestData from '../types/test';
import IDeviceData from '../types/device';

type DeviceContextProps = {
    device: IDeviceData,
    setDevice: (t: IDeviceData) => {},
};

export const DeviceContext = createContext<Partial<DeviceContextProps>>({});
/*
export const DeviceContextProvider = (props: DeviceContextProps): ReactElement => {
    const { children } = props;
    const [ device, setDevice ] = useState<IDeviceData>();
    return (
        <DeviceContext.Provider value={ { device, setDevice }}>
            {children}
        </DeviceContext.Provider>
    );
};
*/