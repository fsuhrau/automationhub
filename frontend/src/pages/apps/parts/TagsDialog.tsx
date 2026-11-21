import React from 'react';
import {Box, Button, Chip, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, Stack, TextField} from '@mui/material';
import {IAppBinaryData} from '../../../types/app';
import {bundleVersion, parseTags} from '../bundlesModel';

export interface TagsEdit {
    bundle: IAppBinaryData,
    text: string,
}

interface TagsDialogProps {
    edit: TagsEdit,
    busy: boolean,
    onChange: (text: string) => void,
    onSave: () => void,
    onClose: () => void,
}

// TagsDialog edits the comma separated tags of a bundle
export const TagsDialog: React.FC<TagsDialogProps> = ({edit, busy, onChange, onSave, onClose}) =>
    <Dialog open={true} onClose={onClose} fullWidth={true} maxWidth="xs" aria-labelledby="tags-title">
        <Box component="form" onSubmit={(e: React.FormEvent) => {
            e.preventDefault();
            onSave();
        }}>
            <DialogTitle id="tags-title">Tags of version {bundleVersion(edit.bundle)}</DialogTitle>
            <DialogContent>
                <DialogContentText variant="body2" sx={{mb: 2}}>
                    Labels to find a bundle by, like a branch or "release candidate". Separate them with commas.
                </DialogContentText>
                <TextField autoFocus={true} fullWidth={true} size="small" label="Tags" placeholder="nightly, rc1" value={edit.text}
                    onChange={e => onChange(e.target.value)}/>
                <Stack direction="row" spacing={0.5} flexWrap="wrap" useFlexGap={true} sx={{mt: 1}}>
                    {parseTags(edit.text).map(t => <Chip key={t} size="small" label={t}/>)}
                </Stack>
            </DialogContent>
            <DialogActions>
                <Button onClick={onClose}>Cancel</Button>
                <Button type="submit" variant="contained" disabled={busy}>Save</Button>
            </DialogActions>
        </Box>
    </Dialog>;
