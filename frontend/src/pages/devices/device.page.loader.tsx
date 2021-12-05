import { FC, useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { Typography } from '@material-ui/core';
import IDeviceData from '../../types/device';
import DeviceEditContent from './device.edit.component';
import { getDevice } from '../../services/device.service';

interface DevicePageProps {
    edit: boolean
}

interface ParamTypes {
    deviceId: string
}

const DevicePageLoader: FC<DevicePageProps> = (props) => {
    const { deviceId } = useParams<ParamTypes>();

    const { edit } = props;
    const [device, setDevice] = useState<IDeviceData>();

    useEffect(() => {
        getDevice(deviceId).then(response => {
            setDevice(response.data);
        }).catch(ex => {
            console.log(ex);
        });
    }, [deviceId]);

    return (
        <div>
            { device
                ? <DeviceEditContent device={ device } />
                : <Typography variant={ 'h1' }>Loading</Typography>
            }
        </div>
    );
};

export default DevicePageLoader;
