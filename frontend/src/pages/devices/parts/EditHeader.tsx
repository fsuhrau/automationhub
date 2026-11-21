import React from 'react';
import {Box, Button, CircularProgress, Stack, Tooltip, Typography} from '@mui/material';
import IDeviceData from '../../../types/device';
import PlatformTypeIcon from '../../../components/PlatformTypeIcon';

interface EditHeaderProps {
    device: IDeviceData,
    dirty: boolean,
    valid: boolean,
    saving: boolean,
    onCancel: () => void,
}

const saveHint = (valid: boolean, dirty: boolean): string => {
    if (!valid) return 'Fix the marked parameters first';
    return dirty ? 'Save (⌘S / Ctrl+S)' : 'Nothing to save yet';
};

// EditHeader: which device is edited, and the way out (cancel or save)
export const EditHeader: React.FC<EditHeaderProps> = ({device, dirty, valid, saving, onCancel}) =>
    <Stack direction={{xs: 'column', sm: 'row'}} spacing={2} useFlexGap={true} alignItems={{sm: 'center'}} sx={{mb: 2}}>
        <Stack direction="row" spacing={1.5} alignItems="center" sx={{minWidth: 0}}>
            <Box sx={{display: 'flex', p: 1, borderRadius: 1, border: 1, borderColor: 'divider'}}>
                <PlatformTypeIcon platformType={device.platformType}/>
            </Box>
            <Box sx={{minWidth: 0}}>
                <Typography component="h1" variant="h5" sx={{wordBreak: 'break-word'}}>Edit {device.alias || device.name}</Typography>
                <Typography variant="body2" color="text.secondary">{dirty ? 'Unsaved changes' : 'No changes yet'}</Typography>
            </Box>
        </Stack>
        <Stack direction="row" spacing={1} sx={{ml: {sm: 'auto'}}}>
            <Button onClick={onCancel} disabled={saving}>Cancel</Button>
            <Tooltip title={saveHint(valid, dirty)}>
                <span>
                    <Button type="submit" variant="contained" disabled={!valid || !dirty || saving}
                        startIcon={saving ? <CircularProgress size={16} color="inherit"/> : undefined}>
                        Save
                    </Button>
                </span>
            </Tooltip>
        </Stack>
    </Stack>;
