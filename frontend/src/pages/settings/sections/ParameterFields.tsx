import React from 'react';
import {Box, Button, Chip, FormControl, InputLabel, MenuItem, Select, Stack, TextField} from '@mui/material';
import {ParameterFormState} from '../useParameterForm';

type FieldsProps = Pick<ParameterFormState, 'form' | 'set' | 'options' | 'errors'>;

export const OptionsEditor: React.FC<ParameterFormState> = ({form, set, option, setOption, addOption, options, errors}) =>
    <Box>
        <Stack direction="row" spacing={0.5} flexWrap="wrap" useFlexGap={true} sx={{mb: 1}}>
            {options.map(o => <Chip key={o} size="small" label={o}
                onDelete={() => set({options: options.filter(x => x !== o), defaultValue: form.defaultValue === o ? '' : form.defaultValue})}/>)}
        </Stack>
        <Stack direction="row" spacing={1} alignItems="flex-start">
            <TextField size="small" label="Option" value={option} onChange={e => setOption(e.target.value)}
                error={!!errors.options} helperText={errors.options ?? 'Enter adds it; commas separate several.'}
                onKeyDown={e => {
                           if (e.key !== 'Enter') return;
                           e.preventDefault();
                           addOption();
                       }}/>
            <Button variant="outlined" onClick={addOption} disabled={!option.trim()}>Add</Button>
        </Stack>
    </Box>;

// DefaultField picks the default from the options of a choice, else types it
export const DefaultField: React.FC<FieldsProps> = ({form, set, options, errors}) => {
    if (form.kind === 'string') {
        return <TextField size="small" label="Default" value={form.defaultValue} onChange={e => set({defaultValue: e.target.value})}
            helperText="Filled in when a test is started; can be changed there."/>;
    }
    return <FormControl size="small" error={!!errors.defaultValue}>
        <InputLabel id="param-default">Default</InputLabel>
        <Select labelId="param-default" label="Default" value={options.includes(form.defaultValue) ? form.defaultValue : ''}
            onChange={e => set({defaultValue: e.target.value})}>
            <MenuItem value=""><em>none</em></MenuItem>
            {options.map(o => <MenuItem key={o} value={o}>{o}</MenuItem>)}
        </Select>
    </FormControl>;
};
