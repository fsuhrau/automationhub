import React, {useState} from 'react';
import {Box, Button, Typography} from '@mui/material';
import {AddRounded} from '@mui/icons-material';
import {ConfirmDialog, Section} from '../../components/common';
import {AppParameter} from '../../types/app';
import AppDetails from './sections/AppDetails';
import AppList from './sections/AppList';
import NewAppDialog from './sections/NewAppDialog';
import ParameterDialog from './sections/ParameterDialog';
import {useAppsSettings} from './useAppsSettings';
import {useConfirmAction} from './useConfirmAction';

// a parameter dialog is open for a parameter, or for a new one (null)
type ParameterTarget = { parameter: AppParameter | null };

// AppsGroup lists the project's apps and edits one: its name and the
// parameters a test run passes to it
const AppsGroup: React.FC = () => {
    const {apps, app, select, rename, removeParameter, createApp, saveParameter} = useAppsSettings();
    const [creating, setCreating] = useState(false);
    const [editing, setEditing] = useState<ParameterTarget | null>(null);
    const deleting = useConfirmAction(removeParameter);

    const addButton = <Button variant="contained" size="small" startIcon={<AddRounded/>} onClick={() => setCreating(true)}>Add app</Button>;

    return <Section title="Apps" description="The builds this project tests, one per platform, with the parameters runs pass to them."
        action={addButton}>
        {apps.length === 0
            ? <Typography variant="body2" color="text.secondary">No apps yet. Add one to create tests and upload bundles.</Typography>
            : <Box sx={{display: 'grid', gridTemplateColumns: {xs: '1fr', md: '260px 1fr'}, gap: 2}}>
                <AppList apps={apps} selectedId={app?.id} onSelect={select}/>
                {app && <AppDetails app={app} onRename={rename.run} onAdd={() => setEditing({parameter: null})}
                    onEdit={parameter => setEditing({parameter})} onDelete={deleting.ask}/>}
            </Box>}
        <NewAppDialog open={creating} onCreate={createApp} onClose={() => setCreating(false)}/>
        <ParameterDialog open={!!editing} parameter={editing?.parameter ?? null} siblings={app?.parameter ?? []}
            onSave={saveParameter} onClose={() => setEditing(null)}/>
        <ConfirmDialog open={!!deleting.target} title={`Delete ${deleting.target?.name}?`} confirm="Delete parameter"
            busy={deleting.busy} onConfirm={deleting.confirm} onClose={deleting.cancel}>
            New runs of {app?.name} won't ask for it or pass it. Runs that already used it keep their values.
        </ConfirmDialog>
    </Section>;
};

export default AppsGroup;
