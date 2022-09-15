import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { Typography } from '@mui/material';
import IDeviceData from '../../types/device';
import DeviceEditContent from './device.edit.content';
import { getDevice } from '../../services/device.service';
import DeviceShowContent from './device.show.content';
import { useProjectContext } from "../../project/project.context";

interface DevicePageProps {
    edit: boolean
}

const DevicePageLoader: React.FC<DevicePageProps> = (props) => {

    const { projectId } = useProjectContext();

    const { deviceId } = useParams();

    const { edit } = props;
    const [device, setDevice] = useState<IDeviceData>();

    useEffect(() => {
        getDevice(deviceId === undefined ? "" : deviceId).then(response => {
            setDevice(response.data);
        }).catch(ex => {
            console.log(ex);
        });
    }, [deviceId]);

    return (
        <div>
            { device
                ? (edit ? <DeviceEditContent device={ device } /> : <DeviceShowContent device={ device } />)
                : <Typography variant={ 'h1' }>Loading</Typography>
            }
        </div>
    );
};

export default DevicePageLoader;
