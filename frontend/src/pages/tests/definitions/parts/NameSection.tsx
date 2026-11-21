import React from 'react';
import {TextField} from '@mui/material';
import {Section} from '../../../../components/common';

interface NameSectionProps {
    name: string,
    onChange: (name: string) => void,
    creating: boolean,
    error?: string,
}

export const NameSection: React.FC<NameSectionProps> = ({name, onChange, creating, error}) =>
    <Section title="Name">
        <TextField size="small" fullWidth={true} required={true} value={name} onChange={e => onChange(e.target.value)}
                   // the section title names the field
            slotProps={{htmlInput: {'aria-label': 'Name'}}}
            autoFocus={creating} placeholder={creating ? 'e.g. Smoke tests' : undefined}
            error={!!error} helperText={error}/>
    </Section>;
