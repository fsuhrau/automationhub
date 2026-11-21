import React from 'react';
import {Alert, Box, Button, Dialog, DialogActions, DialogContent, DialogTitle, Stack, TextField} from '@mui/material';
import {useAsyncAction} from '../../../hooks/useAsyncAction';
import {NewAppRequest} from '../../../types/settings.requests';
import {useNewAppForm} from '../useNewAppForm';
import {IdentifierField, PlatformSelect} from './NewAppFields';

interface NewAppFormProps {
    onCreate: (request: NewAppRequest) => Promise<boolean>,
    onClose: () => void,
}

// the form's state lives here, so every opening of the dialog starts empty
const NewAppForm: React.FC<NewAppFormProps> = ({onCreate, onClose}) => {
    const form = useNewAppForm();
    const action = useAsyncAction(onCreate);
    const submit = async (): Promise<void> => {
        if (form.validate() && await action.run(form.request())) onClose();
    };

    return <Box component="form" noValidate={true} onSubmit={e => {
        e.preventDefault();
        void submit();
    }}>
        <DialogTitle id="new-app-title">Add an app</DialogTitle>
        <DialogContent>
            <Stack spacing={2} sx={{pt: 1}}>
                <TextField size="small" autoFocus={true} label="Name" placeholder="e.g. Game Android" value={form.name}
                    onChange={e => form.setName(e.target.value)} error={!!form.errors.name} helperText={form.errors.name}/>
                <PlatformSelect value={form.platform} onChange={form.setPlatform}/>
                <IdentifierField platform={form.platform} value={form.identifier} error={form.errors.identifier} onChange={form.setIdentifier}/>
                {action.error && <Alert severity="error">{action.error}</Alert>}
            </Stack>
        </DialogContent>
        <DialogActions>
            <Button onClick={onClose}>Cancel</Button>
            <Button type="submit" variant="contained" disabled={action.busy}>Add app</Button>
        </DialogActions>
    </Box>;
};

export interface NewAppDialogProps extends NewAppFormProps {
    open: boolean,
}

const NewAppDialog: React.FC<NewAppDialogProps> = ({open, ...form}) =>
    <Dialog open={open} onClose={form.onClose} fullWidth={true} maxWidth="sm" aria-labelledby="new-app-title">
        <NewAppForm {...form}/>
    </Dialog>;

export default NewAppDialog;
