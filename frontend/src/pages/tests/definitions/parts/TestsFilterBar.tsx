import React from 'react';
import {Chip, InputAdornment, Stack, TextField} from '@mui/material';
import {SearchRounded} from '@mui/icons-material';
import {HealthFilter} from '../../testsListModel';

const FILTERS: { key: HealthFilter, label: string }[] = [
    {key: 'all', label: 'All'},
    {key: 'failing', label: 'Failing'},
    {key: 'passing', label: 'Passing'},
    {key: 'running', label: 'Running'},
    {key: 'never', label: 'Never ran'},
];

interface FilterChipsProps {
    counts: Record<HealthFilter, number>,
    filter: HealthFilter,
    onFilter: (filter: HealthFilter) => void,
}

const FilterChips: React.FC<FilterChipsProps> = ({counts, filter, onFilter}) =>
    <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap={true} role="group" aria-label="Show tests by how they are doing">
        {FILTERS.filter(f => f.key === 'all' || counts[f.key] > 0).map(f => {
            const on = filter === f.key;
            return <Chip key={f.key} size="small" label={`${f.label} ${counts[f.key]}`}
                color={on ? (f.key === 'failing' ? 'error' : 'primary') : 'default'}
                variant={on ? 'filled' : 'outlined'}
                onClick={() => onFilter(on ? 'all' : f.key)}/>;
        })}
    </Stack>;

const SearchField: React.FC<{ value: string, onChange: (text: string) => void }> = ({value, onChange}) =>
    <TextField size="small" placeholder="Find a test" value={value} onChange={e => onChange(e.target.value)}
        sx={{ml: {sm: 'auto'}, width: {xs: '100%', sm: 240}}}
        slotProps={{
                   input: {startAdornment: <InputAdornment position="start"><SearchRounded fontSize="small"/></InputAdornment>},
                   htmlInput: {'aria-label': 'Find a test'},
               }}/>;

interface TestsFilterBarProps extends FilterChipsProps {
    showChips: boolean,
    showSearch: boolean,
    search: string,
    onSearch: (text: string) => void,
}

// TestsFilterBar filters the list by health and searches it by name
export const TestsFilterBar: React.FC<TestsFilterBarProps> = ({showChips, showSearch, search, onSearch, ...chips}) =>
    <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap={true} sx={{mb: 1.5}}>
        {showChips && <FilterChips {...chips}/>}
        {showSearch && <SearchField value={search} onChange={onSearch}/>}
    </Stack>;
