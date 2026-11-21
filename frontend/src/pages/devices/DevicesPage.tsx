import React from 'react';
import {Box} from '@mui/system';
import NodeDevices from '../../components/NodeDevices';
import PageLoader from '../../components/PageLoader';
import {TitleCard} from '../../components/title.card.component';
import {useDevicePool} from './useDevicePool';

// DevicesPage is the device pool of the project, grouped by node. The live
// updates come through the hub state; the page loads the list on entry.
const DevicesPage: React.FC = () => {
    const {error, retry} = useDevicePool();
    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}}}>
            <TitleCard title="Device Pool"
                subtitle="All devices of the project, grouped by the node they are connected to. Open a device for its details, or take live control of it.">
                {error ? <PageLoader error={error} onRetry={retry}/> : <NodeDevices/>}
            </TitleCard>
        </Box>
    );
};

export default DevicesPage;
