import React from 'react';
import {TextField} from '@mui/material';
import {NameField} from '../useNameField';

export interface RequiredNameFieldProps {
    field: NameField,
    placeholder: string,
    hint: string,
    minWidth?: number,
}

// RequiredNameField is the "Name" input of a small form; it complains only
// after a submit was tried
const RequiredNameField: React.FC<RequiredNameFieldProps> = ({field, placeholder, hint, minWidth = 240}) =>
    <TextField size="small" label="Name" placeholder={placeholder} value={field.name} onChange={e => field.setName(e.target.value)}
        error={field.invalid} helperText={field.invalid ? 'A name is needed.' : hint} sx={{minWidth}}/>;

export default RequiredNameField;
