import React from 'react';
import {Alert, Box, Button, LinearProgress, Stack, Typography} from '@mui/material';
import {BundleUpload} from '../../../types/bundle.upload';
import {byteFormat} from '../../tests/value_formatter';
import {isActive, uploadSeverity, uploadText} from '../uploadModel';
import {BundleUploads} from '../useBundleUpload';

interface UploadRowProps {
    upload: BundleUpload,
    onCancel: () => void,
    onDismiss: () => void,
}

const Progress: React.FC<{ upload: BundleUpload }> = ({upload: {state}}) =>
    <LinearProgress sx={{mt: 1, height: 6}} variant={state.phase === 'sending' ? 'determinate' : 'indeterminate'}
        value={state.phase === 'sending' ? state.sent * 100 : undefined}/>;

const UploadRow: React.FC<UploadRowProps> = ({upload, onCancel, onDismiss}) => {
    const {file, state} = upload;
    const active = isActive(state);
    return <Alert severity={uploadSeverity(state)} onClose={active ? undefined : onDismiss} sx={{'& .MuiAlert-message': {flex: 1}}}
        action={active ? <Button color="inherit" size="small" onClick={onCancel}>Cancel</Button> : undefined}>
        <Typography variant="body2" sx={{fontWeight: 600, wordBreak: 'break-all'}}>
            {file.name} <Typography component="span" variant="caption" color="text.secondary">{byteFormat(file.size)}</Typography>
        </Typography>
        <Typography variant="body2">{uploadText(state)}</Typography>
        {active && <Progress upload={upload}/>}
    </Alert>;
};

// UploadQueue shows the files being uploaded and how each one ended
export const UploadQueue: React.FC<{ uploads: BundleUploads }> = ({uploads}) => {
    if (uploads.uploads.length === 0) return null;
    return <Stack spacing={1} sx={{mb: 2}} aria-live="polite">
        {uploads.uploads.map(u => <UploadRow key={u.id} upload={u} onCancel={() => uploads.cancel(u.id)} onDismiss={() => uploads.dismiss(u.id)}/>)}
        {!uploads.busy && uploads.uploads.length > 1 && <Box><Button size="small" onClick={uploads.clear}>Clear all</Button></Box>}
    </Stack>;
};
