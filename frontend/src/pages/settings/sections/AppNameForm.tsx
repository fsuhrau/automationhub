import React, {useState} from 'react';
import {Box, Button, Stack, TextField} from '@mui/material';

export interface AppNameFormProps {
    name: string,
    // resolves true when the name was stored
    onSave: (name: string) => Promise<boolean | undefined>,
}

// AppNameForm renames the app; while the draft is null the field shows the
// stored name, so mounting it per app (key) starts from that app's name
const AppNameForm: React.FC<AppNameFormProps> = ({name, onSave}) => {
    const [draft, setDraft] = useState<string | null>(null);
    const text = draft ?? name;
    const trimmed = text.trim();
    const changed = trimmed !== '' && trimmed !== name;
    const save = async (): Promise<void> => {
        if (changed && await onSave(trimmed)) setDraft(null);
    };

    return <Box component="form" onSubmit={e => {
        e.preventDefault();
        void save();
    }}>
        <Stack direction="row" spacing={1} alignItems="flex-start">
            <TextField size="small" fullWidth={true} label="Name" value={text} onChange={e => setDraft(e.target.value)}
                error={!trimmed} helperText={!trimmed ? 'A name is needed.' : undefined}/>
            {changed && <>
                <Button type="submit" variant="contained">Save</Button>
                <Button onClick={() => setDraft(null)}>Undo</Button>
            </>}
        </Stack>
    </Box>;
};

export default AppNameForm;
