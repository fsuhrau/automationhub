import React from 'react';
import {Chip, InputAdornment, Stack, TextField} from '@mui/material';
import {SearchRounded} from '@mui/icons-material';
import {TestResultState} from '../../../../types/test.result.state.enum';
import {ResultCounts, ResultFilter} from '../../testRunModel';
import {RESULT_ORDER, resultLabel, resultPalette} from '../runStatusModel';

interface ResultFiltersProps {
    counts: ResultCounts,
    total: number,
    running: boolean,
    result: ResultFilter,
    onResult: (r: ResultFilter) => void,
    search: string,
    onSearch: (s: string) => void,
}

const ResultChip: React.FC<{ state: TestResultState, props: ResultFiltersProps }> = ({state, props}) => {
    const {counts, running, result, onResult} = props;
    const on = result === state;
    return <Chip size="small" label={`${resultLabel(state, running)} ${counts[state]}`} disabled={counts[state] === 0}
        color={on ? resultPalette(state) : 'default'} variant={on ? 'filled' : 'outlined'}
        onClick={() => onResult(on ? 'all' : state)}/>;
};

const SearchField: React.FC<{ value: string, onChange: (s: string) => void }> = ({value, onChange}) =>
    <TextField size="small" placeholder="Find a test" value={value} onChange={e => onChange(e.target.value)}
        sx={{ml: {sm: 'auto'}, width: {xs: '100%', sm: 220}}}
        slotProps={{
                   input: {startAdornment: <InputAdornment position="start"><SearchRounded fontSize="small"/></InputAdornment>},
                   htmlInput: {'aria-label': 'Find a test'},
               }}/>;

// the chips filter the test list by result, the field by name
export const ResultFilters: React.FC<ResultFiltersProps> = (props) => {
    const {counts, total, result, onResult, search, onSearch} = props;
    return <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap={true} sx={{mt: 1.5, mb: 1}}>
        <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap={true} role="group" aria-label="Show tests by result">
            <Chip size="small" label={`All ${total}`} color={result === 'all' ? 'primary' : 'default'}
                variant={result === 'all' ? 'filled' : 'outlined'} onClick={() => onResult('all')}/>
            {RESULT_ORDER.filter(r => counts[r] > 0 || r !== TestResultState.TestResultOpen).map(r =>
                <ResultChip key={r} state={r} props={props}/>)}
        </Stack>
        {total > 5 && <SearchField value={search} onChange={onSearch}/>}
    </Stack>;
};
