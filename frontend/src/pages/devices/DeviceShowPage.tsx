import React, {useState} from 'react';
import {Box} from '@mui/material';
import Grid from '@mui/material/Grid';
import IDeviceData from '../../types/device';
import {describeController} from '../../components/DeviceController';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useAuth} from '../../hooks/AuthProvider';
import {SseSourceFactory} from '../../hooks/useSseLog';
import {routes} from '../../lib/routes';
import {availability} from './deviceDetailsModel';
import {useLiveDevice} from './useLiveDevice';
import {useDeviceActions} from './useDeviceActions';
import {useScreenshot} from './useScreenshot';
import {ActionError} from './parts/ActionError';
import {DeviceHeader} from './parts/DeviceHeader';
import {DeviceAlerts} from './parts/DeviceAlerts';
import {DeviceScreen} from './parts/DeviceScreen';
import {DeviceDetailsColumn} from './parts/DeviceDetailsColumn';
import {DeviceConfirm, DeviceDialogs} from './parts/DeviceDialogs';

interface DeviceShowPageProps {
    device: IDeviceData,
    // opens the log stream; tests inject a fake
    createLogSource?: SseSourceFactory,
}

// DeviceShowPage answers whether the device can be used now, what it shows,
// what ran on it lately, and how it is set up.
const DeviceShowPage: React.FC<DeviceShowPageProps> = ({device: stored, createLogSource}) => {
    const {projectIdentifier} = useProjectContext();
    const {device, node} = useLiveDevice(stored);
    const {user} = useAuth();
    const avail = availability(device, describeController, user?.id);
    const links = routes.project(projectIdentifier).device(device.id);
    const shot = useScreenshot(projectIdentifier, device.id);
    const actions = useDeviceActions(device.id);
    const [confirm, setConfirm] = useState<DeviceConfirm | null>(null);

    const onConfirm = async (what: DeviceConfirm): Promise<void> => {
        await (what === 'delete' ? actions.remove() : actions.unlock());
        setConfirm(null);
    };

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}, mt: 2}}>
            <DeviceHeader device={device} node={node} avail={avail} links={links}
                onUnlock={() => setConfirm('unlock')} onDelete={() => setConfirm('delete')}/>
            <ActionError message={actions.error}/>
            <DeviceAlerts device={device} avail={avail} links={links} onUnlock={() => setConfirm('unlock')}/>
            <Grid container={true} spacing={2}>
                <Grid size={{xs: 12, md: 5, lg: 4}}>
                    <DeviceScreen name={device.alias || device.name} avail={avail} links={links} shot={shot}/>
                </Grid>
                <Grid size={{xs: 12, md: 7, lg: 8}}>
                    <DeviceDetailsColumn projectId={projectIdentifier} device={device} node={node} links={links}
                        createLogSource={createLogSource}/>
                </Grid>
            </Grid>
            <DeviceDialogs device={device} confirm={confirm} busy={actions.busy} onConfirm={onConfirm}
                onClose={() => setConfirm(null)}/>
        </Box>
    );
};

export default DeviceShowPage;
