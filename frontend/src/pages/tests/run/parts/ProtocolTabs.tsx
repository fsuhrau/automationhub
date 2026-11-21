import React from 'react';
import {Box, Tab, Tabs} from '@mui/material';
import {TabKey} from '../protocolViewModel';

interface ProtocolTabsProps {
    tab: TabKey,
    onTab: (tab: TabKey) => void,
    counts: { log: number, screenshots: number, performance: number },
}

// the tabs of the protocol page; one without entries is disabled
export const ProtocolTabs: React.FC<ProtocolTabsProps> = ({tab, onTab, counts}) =>
    <Box sx={{borderBottom: 1, borderColor: 'divider', mb: 2}}>
        <Tabs value={tab} onChange={(_, v: TabKey) => onTab(v)} variant="scrollable" allowScrollButtonsMobile={true}>
            <Tab value="overview" label="Overview"/>
            <Tab value="log" label={`Log ${counts.log}`} disabled={counts.log === 0}/>
            <Tab value="screenshots" label={`Screenshots ${counts.screenshots}`} disabled={counts.screenshots === 0}/>
            <Tab value="performance" label={`Performance ${counts.performance}`} disabled={counts.performance === 0}/>
        </Tabs>
    </Box>;
