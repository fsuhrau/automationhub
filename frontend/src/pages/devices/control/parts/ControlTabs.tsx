import React from 'react';
import {Card, CardContent, Tab, Tabs} from '@mui/material';
import {PlatformType} from '../../../../types/platform.type.enum';
import {useSelectedApp} from '../../../../hooks/useSelectedApp';
import {ControlSession} from '../useControlSession';
import {useAppPicker} from '../useAppPicker';
import {useLogs} from '../useLogs';
import {SceneControl} from '../useSceneGraph';
import AppsTab from './AppsTab';
import HierarchyTab from './HierarchyTab';
import LogTab from './LogTab';

export type ControlTab = 'app' | 'inspect' | 'logs';

interface ControlTabsProps {
    session: ControlSession,
    scene: SceneControl,
    tab: ControlTab,
    onTab: (tab: ControlTab) => void,
}

// what can be done beyond the screen: apps, the scene graph and the log. It
// keeps the state of all tabs while another one shows.
const ControlTabs: React.FC<ControlTabsProps> = ({session, scene, tab, onTab}) => {
    const {app: menuApp} = useSelectedApp();
    const platform: PlatformType | undefined = session.device?.platformType;
    const picker = useAppPicker({
        service: session.service, apps: session.apps, platform, menuAppId: menuApp ? menuApp.id as number : null, lease: session.lease.token,
    }, session.run, session.report);
    const logs = useLogs(session.service, session.run);
    return (
        <Card variant="outlined" sx={{flex: '1 1 420px', minWidth: 320}}>
            <Tabs value={tab} onChange={(_, v: ControlTab) => onTab(v)} sx={{px: 1, borderBottom: 1, borderColor: 'divider'}}>
                <Tab value="app" label="App"/>
                <Tab value="inspect" label="Inspect"/>
                <Tab value="logs" label="System log"/>
            </Tabs>
            <CardContent>
                {tab === 'app' && <AppsTab session={session} picker={picker}/>}
                {tab === 'inspect' && <HierarchyTab session={session} control={scene} onAppTab={() => onTab('app')}/>}
                {tab === 'logs' && <LogTab logs={logs} busy={session.busy}/>}
            </CardContent>
        </Card>
    );
};

export default ControlTabs;
