import React from 'react';
import {useParams} from 'react-router-dom';
import DeviceEditPage from './DeviceEditPage';
import DeviceShowPage from './DeviceShowPage';
import {useDeviceService} from '../../services/device.service';
import {useProjectContext} from "../../hooks/ProjectProvider";
import {useLoader} from "../../hooks/useLoader";
import PageLoader from "../../components/PageLoader";

interface DevicePageProps {
    edit: boolean
}

const DevicePageLoader: React.FC<DevicePageProps> = ({edit}) => {
    const {projectIdentifier} = useProjectContext();
    const {deviceId} = useParams();
    const service = useDeviceService();

    const {data: device, error, retry} = useLoader(
        () => deviceId !== undefined ? service.get(projectIdentifier, deviceId) : null,
        [service, projectIdentifier, deviceId]);

    if (!device || error) {
        return <PageLoader error={error} onRetry={retry}/>;
    }
    return edit ? <DeviceEditPage device={device}/> : <DeviceShowPage device={device}/>;
};

export default DevicePageLoader;
