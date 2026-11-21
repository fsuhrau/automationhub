import React from 'react';
import {Box, Button, Stack} from '@mui/material';
import {useNameField} from '../useNameField';
import RequiredNameField from './RequiredNameField';

export interface NodeFormProps {
    busy: boolean,
    // resolves true when the node was added, which empties the form
    onCreate: (name: string) => Promise<boolean | undefined>,
}

const NodeForm: React.FC<NodeFormProps> = ({busy, onCreate}) => {
    const field = useNameField();
    const submit = async (): Promise<void> => {
        if (field.validate() && await onCreate(field.trimmed)) field.reset();
    };

    return <Box component="form" noValidate={true} onSubmit={e => {
        e.preventDefault();
        void submit();
    }} sx={{mb: 2}}>
        <Stack direction={{xs: 'column', sm: 'row'}} spacing={1} alignItems={{sm: 'flex-start'}}>
            <RequiredNameField field={field} placeholder="e.g. Mac mini QA" hint="Shown in the device pool for its devices."/>
            <Button type="submit" variant="contained" disabled={busy}>Add node</Button>
        </Stack>
    </Box>;
};

export default NodeForm;
