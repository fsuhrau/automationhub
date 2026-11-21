import React from 'react';
import {Box, Button, Stack, Typography} from '@mui/material';
import {AddRounded} from '@mui/icons-material';
import ParametersTable, {ParametersTableProps} from './ParametersTable';

export interface ParametersBlockProps extends ParametersTableProps {
    onAdd: () => void,
}

// ParametersBlock lists the parameters a test run passes to the app
const ParametersBlock: React.FC<ParametersBlockProps> = ({onAdd, ...table}) =>
    <Box>
        <Stack direction="row" alignItems="center" sx={{mb: 1}}>
            <Box>
                <Typography variant="body2" sx={{fontWeight: 600}}>Parameters</Typography>
                <Typography variant="caption" color="text.secondary">Asked for when a test of this app is started, and passed to the app.</Typography>
            </Box>
            <Button size="small" startIcon={<AddRounded/>} sx={{ml: 'auto'}} onClick={onAdd}>Add parameter</Button>
        </Stack>
        {table.parameters.length === 0
            ? <Typography variant="body2" color="text.secondary">None.</Typography>
            : <ParametersTable {...table}/>}
    </Box>;

export default ParametersBlock;
