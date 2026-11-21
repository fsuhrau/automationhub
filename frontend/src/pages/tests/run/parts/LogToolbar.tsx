import React from 'react';
import {Button, Chip, InputAdornment, Stack, TextField, Tooltip} from '@mui/material';
import {DownloadRounded, SearchRounded} from '@mui/icons-material';
import {Phase} from '../logModel';
import {LogView} from '../useLogView';
import {ProtocolView} from '../useProtocolView';

const PHASE_HINT: Record<Phase, string> = {
    setup: 'Logged before the test method started: the hub starting the app, OneTimeSetUp and SetUp',
    teardown: 'Logged after the test method ended: TearDown, the hub ending the test, the last screenshot',
};

const SourceChips: React.FC<{ view: ProtocolView, logView: LogView }> = ({view, logView}) =>
    <Stack direction="row" spacing={0.5} flexWrap="wrap" useFlexGap={true} role="group" aria-label="Show sources">
        {view.logSummary.sources.map(s => {
            const on = !logView.hidden.has(s.source);
            return <Chip key={s.source} size="small" label={`${s.source} ${s.count}`}
                variant={on ? 'filled' : 'outlined'} aria-pressed={on} onClick={() => logView.toggleHidden(s.source)}/>;
        })}
    </Stack>;

const PhaseChips: React.FC<{ view: ProtocolView, logView: LogView }> = ({view, logView}) => {
    if (view.phases.length === 0) return null;
    return <Stack direction="row" spacing={0.5} flexWrap="wrap" useFlexGap={true} role="group" aria-label="Show the runner's setup and teardown">
        {view.phases.map(p => {
            const on = !logView.hidden.has(p.phase);
            return <Tooltip key={p.phase} title={PHASE_HINT[p.phase]}>
                <Chip size="small" label={`${p.phase} ${p.count}`} color="secondary"
                    variant={on ? 'filled' : 'outlined'} aria-pressed={on} onClick={() => logView.toggleHidden(p.phase)}/>
            </Tooltip>;
        })}
    </Stack>;
};

const SearchField: React.FC<{ logView: LogView }> = ({logView}) =>
    <TextField size="small" placeholder="Find in log" value={logView.search} onChange={e => logView.setSearch(e.target.value)}
        sx={{width: {xs: '100%', sm: 260}}}
        slotProps={{
                   input: {startAdornment: <InputAdornment position="start"><SearchRounded fontSize="small"/></InputAdornment>},
                   htmlInput: {'aria-label': 'Find in log'},
               }}/>;

interface LogToolbarProps {
    view: ProtocolView,
    logView: LogView,
    onDownload: () => void,
}

// the filters of the log: text, errors, sources and the runner's phases
export const LogToolbar: React.FC<LogToolbarProps> = ({view, logView, onDownload}) => {
    const errors = view.logSummary.errors.length;
    return <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap={true} sx={{mb: 1.5}}>
        <SearchField logView={logView}/>
        <Chip size="small" label={`Errors only ${errors}`} disabled={errors === 0}
            color={logView.errorsOnly ? 'error' : 'default'} variant={logView.errorsOnly ? 'filled' : 'outlined'}
            onClick={() => logView.setErrorsOnly(!logView.errorsOnly)}/>
        <SourceChips view={view} logView={logView}/>
        <PhaseChips view={view} logView={logView}/>
        <Button size="small" startIcon={<DownloadRounded/>} onClick={onDownload} sx={{ml: {sm: 'auto'}}}>Download</Button>
    </Stack>;
};
