import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { Typography } from '@mui/material';
import IDeviceData from '../../types/device';
import DeviceEditContent from './device.edit.content';
import { getDevice } from '../../services/device.service';
import DeviceShowPage from './DeviceShowPage';
import {useProjectContext} from "../../hooks/ProjectProvider";
import {useError} from "../../ErrorProvider";

interface DevicePageProps {
    edit: boolean
}

const DevicePageLoader: React.FC<DevicePageProps> = (props) => {

    const { projectIdentifier } = useProjectContext();

    const { deviceId } = useParams();
    const {setError} = useError()

    const { edit } = props;
    const [device, setDevice] = useState<IDeviceData>();

    useEffect(() => {
        getDevice(projectIdentifier, deviceId === undefined ? "" : deviceId).then(response => {
            setDevice(response.data);
        }).catch(ex => {
            setError(ex);
        });
    }, [projectIdentifier, deviceId]);

    return (
        <div>
            { device
                ? (edit ? <DeviceEditContent device={ device } /> : <DeviceShowPage device={ device } />)
                : <Typography variant={ 'h1' }>Loading</Typography>
            }
        </div>
    );
};

export default DevicePageLoader;
