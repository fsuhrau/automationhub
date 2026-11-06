import React, {useState} from 'react';
import {
    Box,
    Button,
    Card,
    CardContent,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    IconButton,
    Stack,
    Tooltip,
    Typography,
} from '@mui/material';
import {CheckRounded, ContentCopyRounded} from '@mui/icons-material';

// the building blocks of the settings tabs

export const Section: React.FC<{ title: string, description?: React.ReactNode, action?: React.ReactNode, children: React.ReactNode }> =
    ({title, description, action, children}) =>
        <Card variant="outlined">
            <CardContent>
                <Stack direction="row" alignItems="flex-start" spacing={1} useFlexGap sx={{mb: 1.5}}>
                    <Box sx={{minWidth: 0}}>
                        <Typography component="h2" variant="subtitle1" sx={{fontWeight: 600}}>{title}</Typography>
                        {description && <Typography variant="body2" color="text.secondary">{description}</Typography>}
                    </Box>
                    {action && <Box sx={{ml: 'auto', flexShrink: 0}}>{action}</Box>}
                </Stack>
                {children}
            </CardContent>
        </Card>;

export const CopyButton: React.FC<{ text: string, label: string }> = ({text, label}) => {
    const [copied, setCopied] = useState(false);
    return <Tooltip title={copied ? 'Copied' : label}>
        <IconButton size="small" aria-label={label} onClick={() => {
            navigator.clipboard.writeText(text).then(() => {
                setCopied(true);
                setTimeout(() => setCopied(false), 1500);
            }).catch(() => undefined);
        }}>
            {copied ? <CheckRounded fontSize="small" color="success"/> : <ContentCopyRounded fontSize="small"/>}
        </IconButton>
    </Tooltip>;
};

// Code shows text to copy into a terminal or a config file
export const Code: React.FC<{ children: string, label?: string }> = ({children, label = 'Copy'}) =>
    <Box sx={{display: 'flex', alignItems: 'flex-start', gap: 1}}>
        <Box component="pre" sx={{
            m: 0, p: 1.5, flex: 1, overflowX: 'auto', fontSize: 13, whiteSpace: 'pre',
            bgcolor: 'background.default', border: 1, borderColor: 'divider', borderRadius: 1,
        }}>{children}</Box>
        <CopyButton text={children} label={label}/>
    </Box>;

// ConfirmDialog asks before something that can't be undone and says what happens
export const ConfirmDialog: React.FC<{
    open: boolean,
    title: string,
    children: React.ReactNode,
    confirm: string,
    busy?: boolean,
    color?: 'error' | 'warning',
    onConfirm: () => void,
    onClose: () => void,
}> = ({open, title, children, confirm, busy, color = 'error', onConfirm, onClose}) =>
    <Dialog open={open} onClose={onClose} aria-labelledby="confirm-title">
        <DialogTitle id="confirm-title">{title}</DialogTitle>
        <DialogContent><DialogContentText component="div">{children}</DialogContentText></DialogContent>
        <DialogActions>
            <Button onClick={onClose}>Cancel</Button>
            <Button color={color} variant="contained" onClick={onConfirm} disabled={busy}>{confirm}</Button>
        </DialogActions>
    </Dialog>;

export const messageOf = (e: unknown): string => {
    const err = e as { response?: { data?: { message?: string } }, message?: string };
    return err.response?.data?.message ?? err.message ?? String(e);
};
