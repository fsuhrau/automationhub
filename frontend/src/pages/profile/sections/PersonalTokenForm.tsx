import React, {useState} from 'react';
import {Box, Button, Stack} from '@mui/material';
import dayjs, {Dayjs} from 'dayjs';
import ExpiryPicker from '../../settings/sections/ExpiryPicker';
import RequiredNameField from '../../settings/sections/RequiredNameField';
import {useNameField} from '../../settings/useNameField';

export interface PersonalTokenFormProps {
    busy: boolean,
    // resolves true when the token was created, which empties the form
    onCreate: (name: string, expiresAt: Dayjs | null) => Promise<boolean | undefined>,
}

const PersonalTokenForm: React.FC<PersonalTokenFormProps> = ({busy, onCreate}) => {
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
            <RequiredNameField field={field} placeholder="e.g. claude-code" hint="Says where the token is used." minWidth={220}/>
            <ExpiryPicker value={expiresAt} onChange={setExpiresAt} minDate={dayjs().add(1, 'day')}/>
            <Button type="submit" variant="contained" disabled={busy}>Create token</Button>
        </Stack>
    </Box>;
};

export default PersonalTokenForm;
