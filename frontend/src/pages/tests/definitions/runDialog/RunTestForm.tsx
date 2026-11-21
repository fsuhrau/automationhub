import React from 'react';
import {Alert, Box, Button, Dialog, DialogActions, DialogContent, DialogTitle, Stack, TextField} from '@mui/material';
import {PlayArrowRounded} from '@mui/icons-material';
import {IAppData} from '../../../../types/app';
import {acceptFor} from '../../../apps/bundlesModel';
import {BundleField} from './BundleField';
import {ParameterFields} from './ParameterFields';
import {useRunTestForm} from './useRunTestForm';

interface RunTestFormProps {
    projectIdentifier: string,
    app: IAppData,
    test: { id: number, name: string },
    onClose: () => void,
}

// RunTestForm is the open run dialog: a fresh form each time it opens
export const RunTestForm: React.FC<RunTestFormProps> = props => {
    const {app, test, onClose} = props;
    const form = useRunTestForm(props, onClose);
    return (
        <Dialog open={true} onClose={form.starting ? undefined : onClose} fullWidth={true} maxWidth="sm" aria-labelledby="run-test-title">
            <Box component="form" onSubmit={e => {
                e.preventDefault();
                if (form.ready) form.start();
            }}>
                <DialogTitle id="run-test-title">Run {test.name}</DialogTitle>
                <DialogContent>
                    <Stack spacing={2.5} sx={{pt: 1}}>
                        {form.needsBundle && <BundleField bundles={form.bundles} accept={acceptFor(app.platform)}/>}
                        {form.needsUrl && <TextField size="small" required={true} label="Start URL" placeholder="https://" value={form.startUrl}
                            onChange={e => form.setStartUrl(e.target.value)} helperText="The page the browser opens for the test."/>}
                        {app.parameter.length > 0 && <ParameterFields parameters={app.parameter} values={form.params} onChange={form.setParam}/>}
                        {form.error && <Alert severity="error">{form.error}</Alert>}
                    </Stack>
                </DialogContent>
                <DialogActions>
                    <Button onClick={onClose} disabled={form.starting}>Cancel</Button>
                    <Button type="submit" variant="contained" startIcon={<PlayArrowRounded/>} disabled={!form.ready}>
                        {form.starting ? 'Starting…' : 'Start run'}
                    </Button>
                </DialogActions>
            </Box>
        </Dialog>
    );
};
