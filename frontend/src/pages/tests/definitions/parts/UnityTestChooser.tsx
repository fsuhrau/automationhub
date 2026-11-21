import React, {useMemo, useState} from 'react';
import {Box, Button, Checkbox, FormControlLabel, InputAdornment, Stack, TextField, Typography} from '@mui/material';
import {SearchRounded} from '@mui/icons-material';
import IAppFunctionData from '../../../../types/app.function';
import {functionKey, groupByClass, matchesFunction, noChoicesText, shortClass, shortMethod, toggled} from '../../testEditModel';

interface ChoiceListProps {
    choices: IAppFunctionData[],
    checked: Set<string>,
    onToggle: (key: string, on: boolean) => void,
}

const ChoiceList: React.FC<ChoiceListProps> = ({choices, checked, onToggle}) =>
    <Box sx={{maxHeight: 320, overflowY: 'auto', border: 1, borderColor: 'divider', borderRadius: 1, p: 1}}
        role="group" aria-label="Tests to add">
        {groupByClass(choices).map(([cls, fns]) => <Box key={cls} sx={{mb: 1}}>
            <Typography variant="caption" color="text.secondary" title={cls}>{shortClass(cls)}</Typography>
            {fns.map(f => <FormControlLabel key={functionKey(f)} sx={{display: 'flex', pl: 0.5, mr: 0}}
                control={<Checkbox size="small" checked={checked.has(functionKey(f))}
                    onChange={(_, on) => onToggle(functionKey(f), on)}/>}
                label={<Typography variant="body2" sx={{fontFamily: 'monospace', fontSize: 13}}>{shortMethod(f.method)}</Typography>}/>)}
        </Box>)}
    </Box>;

interface UnityTestChooserProps {
    available: IAppFunctionData[],
    picked: IAppFunctionData[],
    onAdd: (functions: IAppFunctionData[]) => void,
}

// UnityTestChooser searches the tests Unity has and adds the checked ones
// or all that were found
export const UnityTestChooser: React.FC<UnityTestChooserProps> = ({available, picked, onAdd}) => {
    const [search, setSearch] = useState('');
    const [checked, setChecked] = useState<Set<string>>(new Set());
    const pickedKeys = useMemo(() => new Set(picked.map(functionKey)), [picked]);
    const choices = available.filter(f => !pickedKeys.has(functionKey(f)) && matchesFunction(f, search));
    const add = (list: IAppFunctionData[]): void => {
        onAdd(list);
        setChecked(new Set());
    };
    return <Stack spacing={1}>
        <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap={true}>
            <TextField size="small" placeholder="Find a test or class" value={search} onChange={e => setSearch(e.target.value)}
                sx={{flex: 1, minWidth: 200}}
                slotProps={{
                           input: {startAdornment: <InputAdornment position="start"><SearchRounded fontSize="small"/></InputAdornment>},
                           htmlInput: {'aria-label': 'Find a test to add'},
                       }}/>
            <Button size="small" variant="contained" disabled={checked.size === 0}
                onClick={() => add(choices.filter(f => checked.has(functionKey(f))))}>Add {checked.size || ''} picked</Button>
            <Button size="small" disabled={choices.length === 0} onClick={() => add(choices)}>
                Add all {choices.length}{search ? ' found' : ''}
            </Button>
        </Stack>
        {choices.length === 0
            ? <Typography variant="body2" color="text.secondary">{noChoicesText(available.length, search)}</Typography>
            : <ChoiceList choices={choices} checked={checked} onToggle={(key, on) => setChecked(prev => toggled(prev, key, on))}/>}
    </Stack>;
};
