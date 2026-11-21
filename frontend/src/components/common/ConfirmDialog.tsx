import React from 'react';
import {Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle} from '@mui/material';

export interface ConfirmDialogProps {
    open: boolean,
    title: string,
    children: React.ReactNode,
    confirm: string,
    busy?: boolean,
    color?: 'error' | 'warning',
    onConfirm: () => void,
    onClose: () => void,
}

// ConfirmDialog asks before something that can't be undone and says what happens
export const ConfirmDialog: React.FC<ConfirmDialogProps> = ({open, title, children, confirm, busy, color = 'error', onConfirm, onClose}) =>
    <Dialog open={open} onClose={onClose} aria-labelledby="confirm-title">
        <DialogTitle id="confirm-title">{title}</DialogTitle>
        <DialogContent><DialogContentText component="div">{children}</DialogContentText></DialogContent>
        <DialogActions>
            <Button onClick={onClose}>Cancel</Button>
            <Button color={color} variant="contained" onClick={onConfirm} disabled={busy}>{confirm}</Button>
        </DialogActions>
    </Dialog>;
