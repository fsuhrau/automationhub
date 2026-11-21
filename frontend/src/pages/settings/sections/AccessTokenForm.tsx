import React, {useState} from 'react';
import {Box, Button, Stack} from '@mui/material';
import {Dayjs} from 'dayjs';
import {useNameField} from '../useNameField';
import ExpiryPicker from './ExpiryPicker';
import RequiredNameField from './RequiredNameField';

export interface AccessTokenFormProps {
    busy: boolean,
    // resolves true when the token was created, which empties the form
    onCreate: (name: string, expiresAt: Dayjs | null) => Promise<boolean | undefined>,
}

const AccessTokenForm: React.FC<AccessTokenFormProps> = ({busy, onCreate}) => {
    const field = useNameField();
    const [expiresAt, setExpiresAt] = useState<Dayjs | null>(null);
    const submit = async (): Promise<void> => {
        if (!field.validate() || !await onCreate(field.trimmed, expiresAt)) return;
        field.reset();
        setExpiresAt(null);
    };

    return <Box component="form" noValidate={true} onSubmit={e => {
        e.preventDefault();
        void submit();
    }} sx={{mb: 2}}>
        <Stack direction={{xs: 'column', sm: 'row'}} spacing={1} alignItems={{sm: 'flex-start'}}>
            <RequiredNameField field={field} placeholder="e.g. CI nightly" hint="Says where the token is used."/>
            <ExpiryPicker value={expiresAt} onChange={setExpiresAt}/>
            <Button type="submit" variant="contained" disabled={busy}>Create token</Button>
        </Stack>
    </Box>;
};

export default AccessTokenForm;
