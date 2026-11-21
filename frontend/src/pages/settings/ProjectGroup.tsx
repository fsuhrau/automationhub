import React from 'react';
import {Box, Button, Stack, TextField, Typography} from '@mui/material';
import {CopyButton, Section} from '../../components/common';
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useProjectName} from './useProjectName';

// ProjectGroup renames the project and shows its identifier
const ProjectGroup: React.FC = () => {
    const {projectIdentifier} = useProjectContext();
    const {name, setName, changed, saving, save, undo} = useProjectName();
    const empty = name.trim() === '';

    return <Section title="Project" description="How the project is called, and how tools address it.">
        <Stack spacing={2.5} sx={{maxWidth: 640}}>
            <Box component="form" onSubmit={e => {
                e.preventDefault();
                save();
            }}>
                <Stack direction="row" spacing={1} alignItems="flex-start">
                    <TextField size="small" fullWidth={true} label="Name" value={name} onChange={e => setName(e.target.value)}
                        error={empty} helperText={empty ? 'A name is needed.' : 'Shown in the menu and on the project list.'}/>
                    {changed && <Button type="submit" variant="contained" disabled={saving}>Save</Button>}
                    {changed && <Button onClick={undo}>Undo</Button>}
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
