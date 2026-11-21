import React, {useState} from 'react';
import {Alert, Box, Button, Dialog, DialogActions, DialogContent, DialogTitle, TextField, Typography} from '@mui/material';
import {useAsyncAction} from '../../hooks/useAsyncAction';
import {projectSlug} from '../projectsModel';

interface NewProjectFormProps {
    onCreate: (name: string) => Promise<boolean>,
    onClose: () => void,
}

const nameHint = (name: string, tried: boolean): React.ReactNode => {
    if (!name.trim()) return tried ? 'A name is needed.' : ' ';
    return <>Identifier: <b>{projectSlug(name)}</b> (in URLs and the API; it can't change)</>;
};

// the form's state lives here, so every opening of the dialog starts empty
const NewProjectForm: React.FC<NewProjectFormProps> = ({onCreate, onClose}) => {
    const [name, setName] = useState('');
    const [tried, setTried] = useState(false);
    const action = useAsyncAction(onCreate);
    const submit = async (): Promise<void> => {
        setTried(true);
        if (name.trim() && await action.run(name.trim())) onClose();
    };

    return <Box component="form" noValidate={true} onSubmit={e => {
        e.preventDefault();
        void submit();
    }}>
        <DialogTitle id="new-project-title">New project</DialogTitle>
        <DialogContent>
            <Typography variant="body2" color="text.secondary" sx={{mb: 2}}>
                A project holds a game's apps, tests and results. Devices and nodes are shared by all projects.
            </Typography>
            <TextField autoFocus={true} fullWidth={true} size="small" label="Name" placeholder="e.g. Heroes of History" value={name}
                onChange={e => setName(e.target.value)} error={tried && !name.trim()} helperText={nameHint(name, tried)}/>
            {action.error && <Alert severity="error" sx={{mt: 2}}>{action.error}</Alert>}
        </DialogContent>
        <DialogActions>
            <Button onClick={onClose}>Cancel</Button>
            <Button type="submit" variant="contained" disabled={action.busy}>Create project</Button>
        </DialogActions>
    </Box>;
};

export interface NewProjectDialogProps extends NewProjectFormProps {
    open: boolean,
}

const NewProjectDialog: React.FC<NewProjectDialogProps> = ({open, ...form}) =>
    <Dialog open={open} onClose={form.onClose} fullWidth={true} maxWidth="xs" aria-labelledby="new-project-title">
        <NewProjectForm {...form}/>
    </Dialog>;

export default NewProjectDialog;
