import React from 'react';
import {Box, TextField} from '@mui/material';
import {Section} from '../../../components/common';

interface AliasSectionProps {
    name: string,
    alias: string,
    onChange: (alias: string) => void,
}

// AliasSection edits the alias that is shown instead of the device name
export const AliasSection: React.FC<AliasSectionProps> = ({name, alias, onChange}) =>
    <Section title="Alias" description={<>
        Shown instead of the device name <Box component="span" sx={{fontFamily: 'monospace'}}>{name}</Box> in
        lists, test results and the device pool. Leave it empty to use the device name.
    </>}>
        <TextField size="small" fullWidth={true} label="Alias" placeholder={name} value={alias}
            onChange={e => onChange(e.target.value)} slotProps={{htmlInput: {maxLength: 100}}}/>
    </Section>;
