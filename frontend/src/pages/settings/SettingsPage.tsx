import React from 'react';
import {Box, Tab, Tabs, Typography} from '@mui/material';
import {useSearchParams} from 'react-router-dom';
import AccessTokenTab from './AccessTokenTab';
import NodesTab from './NodesTab';
import AppsGroup from './AppsGroup';
import ProjectGroup from './ProjectGroup';
import UsersTab from './UsersTab';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {parseTab, SETTINGS_TABS, TabKey, withTab} from './settingsTabs';

const TAB_VIEWS: Record<TabKey, React.FC> = {
    project: ProjectGroup,
    apps: AppsGroup,
    tokens: AccessTokenTab,
    nodes: NodesTab,
    people: UsersTab,
};

// SettingsPage sets up the project; the tab is in the URL, so a link can
// lead straight to it (e.g. "add an app" from the menu)
const SettingsPage: React.FC = () => {
    const {project} = useProjectContext();
    const [query, setQuery] = useSearchParams();
    const tab = parseTab(query.get('tab'));
    const setTab = (t: TabKey): void => setQuery(q => withTab(q, t), {replace: true});
    const View = TAB_VIEWS[tab];

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}, mt: 2}}>
            <Typography component="h1" variant="h5">Settings</Typography>
            <Typography variant="body2" color="text.secondary">
                How {project.name} is set up: its apps, the tokens tools use, and the nodes that connect devices.
            </Typography>
            <Box sx={{borderBottom: 1, borderColor: 'divider', mt: 1.5, mb: 2}}>
                <Tabs value={tab} onChange={(_, v: TabKey) => setTab(v)} variant="scrollable" allowScrollButtonsMobile={true}>
                    {SETTINGS_TABS.map(t => <Tab key={t.key} value={t.key} label={t.label}/>)}
                </Tabs>
            </Box>
            <Box sx={{maxWidth: 1100}}><View/></Box>
        </Box>
    );
};

export default SettingsPage;
