import React, {useState} from 'react';
import {Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, TextField, Typography} from '@mui/material';
import {DeviceTestRequest} from '../../../types/device.requests';

interface RunOnDeviceDialogProps {
    onStart: (request: DeviceTestRequest) => void,
    onClose: () => void,
}

// RunOnDeviceDialog asks which test to start on a device and with which
// parameters; it is open while it is mounted
export const RunOnDeviceDialog: React.FC<RunOnDeviceDialogProps> = ({onStart, onClose}) => {
    const [testName, setTestName] = useState('');
    const [env, setEnv] = useState('');
    return <Dialog open={true} onClose={onClose} aria-labelledby="form-dialog-title">
        <DialogTitle id="form-dialog-title">Run Test</DialogTitle>
        <DialogContent>
            <DialogContentText>Enter Test to execute</DialogContentText>
            <Typography variant="subtitle1">Parameter:</Typography>
            <Typography variant="subtitle2">
                server=http://localhost:8080<br/>
                user=autohub
            </Typography>
            <br/>
            <TextField placeholder="Parameter" fullWidth={true} multiline={true} rows={4} variant="outlined"
                onChange={e => setEnv(e.target.value)} slotProps={{htmlInput: {'aria-label': 'Parameters'}}}/>
            <Typography variant="subtitle1">Test:</Typography>
            <br/>
            <TextField placeholder="Parameter" fullWidth={true} variant="outlined"
                onChange={e => setTestName(e.target.value)} slotProps={{htmlInput: {'aria-label': 'Test name'}}}/>
        </DialogContent>
        <DialogActions>
            <Button onClick={onClose} color="primary">Cancel</Button>
            <Button onClick={() => onStart({testName, env})} color="primary" variant="contained">Start</Button>
        </DialogActions>
    </Dialog>;
};
