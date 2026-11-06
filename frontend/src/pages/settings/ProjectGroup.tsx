import React, {useEffect, useState} from 'react';
import {Box, Button, Stack, TextField, Typography} from '@mui/material';
import {updateProject} from '../../project/project.service';
import IProject from '../../project/project';
import {HubStateActions} from '../../application/HubState';
import {useHubState} from '../../hooks/HubStateProvider';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useError} from '../../ErrorProvider';
import {CopyButton, messageOf, Section} from './parts';

// ProjectGroup renames the project and shows its identifier
const ProjectGroup: React.FC = () => {
    const {dispatch} = useHubState();
    const {project, projectIdentifier} = useProjectContext();
    const {setError} = useError();
    const [name, setName] = useState(project.name);
    const [saving, setSaving] = useState(false);
    useEffect(() => setName(project.name), [project.name]);

    const changed = name.trim() !== project.name && name.trim() !== '';
    const save = () => {
        if (!changed) return;
        setSaving(true);
        updateProject(projectIdentifier, {...project, name: name.trim()} as IProject).then(() => {
            dispatch({type: HubStateActions.ProjectAttributeUpdate, payload: {projectIdentifier, attribute: 'name', value: name.trim()}});
        }).catch(e => setError(messageOf(e))).finally(() => setSaving(false));
    };

    return <Section title="Project" description="How the project is called, and how tools address it.">
        <Stack spacing={2.5} sx={{maxWidth: 640}}>
            <Box component="form" onSubmit={e => {
                e.preventDefault();
                save();
            }}>
                <Stack direction="row" spacing={1} alignItems="flex-start">
                    <TextField size="small" fullWidth label="Name" value={name} onChange={e => setName(e.target.value)}
                               error={name.trim() === ''} helperText={name.trim() === '' ? 'A name is needed.' : 'Shown in the menu and on the project list.'}/>
                    {changed && <Button type="submit" variant="contained" disabled={saving}>Save</Button>}
                    {changed && <Button onClick={() => setName(project.name)}>Undo</Button>}
                </Stack>
            </Box>
            <Box>
                <Typography variant="caption" color="text.secondary">Identifier</Typography>
                <Stack direction="row" alignItems="center" spacing={0.5}>
                    <Typography variant="body2" sx={{fontFamily: 'monospace'}}>{projectIdentifier}</Typography>
                    <CopyButton text={projectIdentifier} label="Copy the identifier"/>
                </Stack>
                <Typography variant="caption" color="text.secondary">
                    Part of the project's URLs, the API and the MCP endpoint, so it doesn't change.
                </Typography>
            </Box>
        </Stack>
    </Section>;
};

export default ProjectGroup;
