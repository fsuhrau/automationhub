import React, {useState} from 'react';
import {Typography} from '@mui/material';
import {useNavigate} from 'react-router-dom';
import {TitleCard} from './title.card.component';
import DevicesTable from './DevicesTable';
import {useProjectContext} from '../hooks/ProjectProvider';
import {useHubState} from '../hooks/HubStateProvider';
import {routes} from '../lib/routes';
import {DeviceTestRequest} from '../types/device.requests';
import {filterByPlatform, groupDevices, PlatformChoice, platformCounts, platformsOf} from '../pages/devices/deviceListModel';
import {useDeviceListActions} from '../pages/devices/useDeviceListActions';
import {ActionError} from '../pages/devices/parts/ActionError';
import {PlatformFilter} from '../pages/devices/parts/PlatformFilter';
import {RunOnDeviceDialog} from '../pages/devices/parts/RunOnDeviceDialog';

const NoDevices: React.FC = () =>
    <Typography variant="body2" color="text.secondary" sx={{p: 2}}>
        No devices yet. Devices show up here when a node or the master detects them
        (for example an Android phone connected over USB).
    </Typography>;

// NodeDevices lists the device pool grouped by node, with a platform filter,
// unlock and starting a test on a device. The pool belongs to the project,
// not to an app, so the filter is on the page instead of following the app
// picked in the menu.
const NodeDevices: React.FC = () => {
    const {projectIdentifier} = useProjectContext();
    const navigate = useNavigate();
    const {state} = useHubState();
    const actions = useDeviceListActions();
    const [platform, setPlatform] = useState<PlatformChoice>('all');
    // the device to start a test on, while the dialog is open
    const [runOn, setRunOn] = useState<number | null>(null);

    const all = state.devices ?? [];
    const counts = platformCounts(all);
    const start = (request: DeviceTestRequest): void => {
        void actions.runTest(runOn as number, request);
        setRunOn(null);
    };
    const open = (id: number): void => {
        void navigate(routes.project(projectIdentifier).device(id).show());
    };

    return <>
        {runOn !== null && <RunOnDeviceDialog onStart={start} onClose={() => setRunOn(null)}/>}
        <ActionError message={actions.error}/>
        <PlatformFilter total={all.length} counts={counts} platforms={platformsOf(counts)} selected={platform} onSelect={setPlatform}/>
        {state.devices !== null && all.length === 0 && <NoDevices/>}
        {groupDevices(filterByPlatform(all, platform), state.nodes).map(group =>
            <TitleCard key={group.title} title={group.title}>
                <DevicesTable devices={group.devices} onOpenDeviceDetails={open} onSelectForRun={setRunOn}
                    onUnlockDevice={id => void actions.unlock(id)}/>
            </TitleCard>)}
    </>;
};

export default NodeDevices;
