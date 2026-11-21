import React from 'react';
import {Box, FormControl, InputLabel, MenuItem, Select, Stack, TextField, Typography} from '@mui/material';
import {AppParameter, AppParameterOption} from '../../../../types/app';

interface ParameterFieldsProps {
    parameters: AppParameter[],
    values: Record<string, string>,
    onChange: (name: string, value: string) => void,
}

type FieldProps = Omit<ParameterFieldsProps, 'parameters'> & { parameter: AppParameter };

const OptionField: React.FC<FieldProps> = ({parameter: p, values, onChange}) =>
    <FormControl size="small" fullWidth={true}>
        <InputLabel id={`run-param-${p.id}`}>{p.name}</InputLabel>
        <Select labelId={`run-param-${p.id}`} label={p.name} value={values[p.name] ?? ''} onChange={e => onChange(p.name, e.target.value)}>
            {(p.type as AppParameterOption).options.map(o => <MenuItem key={o} value={o}>
                {o}{o === p.type.defaultValue ? ' (default)' : ''}
            </MenuItem>)}
        </Select>
    </FormControl>;

const TextParameter: React.FC<FieldProps> = ({parameter: p, values, onChange}) =>
    <TextField size="small" fullWidth={true} label={p.name} value={values[p.name] ?? ''}
        placeholder={p.type.defaultValue || undefined} onChange={e => onChange(p.name, e.target.value)}/>;

// ParameterFields has a field for each parameter of the app: a choice for
// options, text for the rest
export const ParameterFields: React.FC<ParameterFieldsProps> = ({parameters, ...rest}) =>
    <Box>
        <Typography variant="subtitle2" sx={{mb: 1}}>Parameters</Typography>
        <Stack spacing={1.5}>
            {parameters.map(p => p.type.type === 'option'
                ? <OptionField key={p.id} parameter={p} {...rest}/>
                : <TextParameter key={p.id} parameter={p} {...rest}/>)}
        </Stack>
        <Typography variant="caption" color="text.secondary">Empty parameters aren't passed.</Typography>
    </Box>;
