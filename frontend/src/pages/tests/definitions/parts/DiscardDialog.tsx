import React from 'react';
import {Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle} from '@mui/material';

interface DiscardDialogProps {
    open: boolean,
    // the name of the test being edited; none for a new test
    testName?: string,
    onKeep: () => void,
    onDiscard: () => void,
}

// DiscardDialog asks before unsaved changes are thrown away
export const DiscardDialog: React.FC<DiscardDialogProps> = ({open, testName, onKeep, onDiscard}) =>
    <Dialog open={open} onClose={onKeep} aria-labelledby="discard-test-title">
        <DialogTitle id="discard-test-title">Discard your changes?</DialogTitle>
        <DialogContent>
            <DialogContentText>{testName ? `The changes to ${testName} aren't saved yet.` : "The new test isn't created yet."}</DialogContentText>
        </DialogContent>
        <DialogActions>
            <Button onClick={onKeep}>Keep editing</Button>
            <Button color="warning" onClick={onDiscard}>Discard</Button>
        </DialogActions>
    </Dialog>;
