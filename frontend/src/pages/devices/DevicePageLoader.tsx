import React from 'react';
import {useParams} from 'react-router-dom';
import DeviceEditPage from './DeviceEditPage';
import {getDevice} from '../../services/device.service';
import DeviceShowPage from './DeviceShowPage';
import {useProjectContext} from "../../hooks/ProjectProvider";
import {useLoader} from "../../hooks/useLoader";
import PageLoader from "../../components/PageLoader";

interface DevicePageProps {
    edit: boolean
}

const DevicePageLoader: React.FC<DevicePageProps> = ({edit}) => {
    const {projectIdentifier} = useProjectContext();
    const {deviceId} = useParams();

    const {data: device, error, retry} = useLoader(
        () => deviceId !== undefined ? getDevice(projectIdentifier, deviceId) : null,
        [projectIdentifier, deviceId]);

    if (!device || error) {
        return <PageLoader error={error} onRetry={retry}/>;
    }
    return edit ? <DeviceEditPage device={device}/> : <DeviceShowPage device={device}/>;
};

export default DevicePageLoader;
