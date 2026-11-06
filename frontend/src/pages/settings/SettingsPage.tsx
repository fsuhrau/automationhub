import React from 'react';
import {Box, Tab, Tabs, Typography} from '@mui/material';
import {useSearchParams} from 'react-router-dom';
import AccessTokenTab from './AccessTokenTab';
import NodesTab from './NodesTab';
import AppsGroup from './AppsGroup';
import ProjectGroup from './ProjectGroup';
import UsersTab from './UsersTab';
import {useProjectContext} from '../../hooks/ProjectProvider';

const TABS = [
    {key: 'project', label: 'Project'},
    {key: 'apps', label: 'Apps'},
    {key: 'tokens', label: 'Access tokens'},
    {key: 'nodes', label: 'Nodes'},
    {key: 'people', label: 'People'},
] as const;
type TabKey = typeof TABS[number]['key'];

// SettingsPage sets up the project; the tab is in the URL, so a link can
// lead straight to it (e.g. "add an app" from the menu)
const SettingsPage: React.FC = () => {
    const {project} = useProjectContext();
    const [query, setQuery] = useSearchParams();
    const tab: TabKey = TABS.some(t => t.key === query.get('tab')) ? query.get('tab') as TabKey : 'project';
    const setTab = (t: TabKey) => setQuery(q => {
        const next = new URLSearchParams(q);
        if (t === 'project') next.delete('tab'); else next.set('tab', t);
        return next;
    }, {replace: true});

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}, mt: 2}}>
            <Typography component="h1" variant="h5">Settings</Typography>
            <Typography variant="body2" color="text.secondary">How {project.name} is set up: its apps, the tokens tools use, and the nodes that connect devices.</Typography>
            <Box sx={{borderBottom: 1, borderColor: 'divider', mt: 1.5, mb: 2}}>
                <Tabs value={tab} onChange={(_, v) => setTab(v)} variant="scrollable" allowScrollButtonsMobile>
                    {TABS.map(t => <Tab key={t.key} value={t.key} label={t.label}/>)}
                </Tabs>
            </Box>
            <Box sx={{maxWidth: 1100}}>
                {tab === 'project' && <ProjectGroup/>}
                {tab === 'apps' && <AppsGroup/>}
                {tab === 'tokens' && <AccessTokenTab/>}
                {tab === 'nodes' && <NodesTab/>}
                {tab === 'people' && <UsersTab/>}
            </Box>
        </Box>
    );
};

export default SettingsPage;
