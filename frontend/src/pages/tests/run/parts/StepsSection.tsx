import React from 'react';
import {Box, Typography} from '@mui/material';
import {ResultIcon, Section} from '../../../../components/common';
import IProtocolEntryData from '../../../../types/protocol.entry';
import {formatRuntime} from '../logModel';
import {resultLabel} from '../runStatusModel';
import {ProtocolView} from '../useProtocolView';
import {Message} from './Message';

const ROW = {display: 'grid', gridTemplateColumns: '80px 16px 1fr', gap: 1} as const;

const Time: React.FC<{ runtime: number | null }> = ({runtime}) =>
    <Typography variant="caption" color="text.secondary" sx={{fontFamily: 'monospace', pt: '2px'}}>
        {runtime === null ? '' : formatRuntime(runtime)}
    </Typography>;

const StepDot: React.FC<{ last: boolean, failed: boolean }> = ({last, failed}) =>
    <Box sx={{position: 'relative', height: '100%', minHeight: 28}}>
        <Box sx={{position: 'absolute', left: 5, top: 7, width: 7, height: 7, borderRadius: '50%',
            bgcolor: last && failed ? 'error.main' : 'text.secondary'}}/>
        {!last && <Box sx={{position: 'absolute', left: 8, top: 16, bottom: -4, width: '1px', bgcolor: 'divider'}}/>}
    </Box>;

const Step: React.FC<{ step: IProtocolEntryData, last: boolean, failed: boolean }> = ({step, last, failed}) =>
    <Box component="li" sx={{...ROW, alignItems: 'start'}}>
        <Time runtime={step.runtime}/>
        <StepDot last={last} failed={failed}/>
        <Box sx={{pb: 1}}><Message text={step.message}/></Box>
    </Box>;

// the last line of a finished test: its result
const Outcome: React.FC<{ view: ProtocolView }> = ({view}) => {
    const {entries, testResult} = view;
    return <Box component="li" sx={ROW}>
        <Time runtime={entries.length > 0 ? entries[entries.length - 1].runtime : null}/>
        <Box sx={{pt: '2px'}}><ResultIcon result={testResult}/></Box>
        <Typography variant="body2" sx={{fontWeight: 600}}>{resultLabel(testResult, true)}</Typography>
    </Box>;
};

// what the test did, step by step, and how it ended
export const StepsSection: React.FC<{ view: ProtocolView }> = ({view}) => {
    const {steps} = view.summary;
    return <Section title="Steps" action={steps.length > 0 && <Typography variant="body2" color="text.secondary">{steps.length}</Typography>}>
        {steps.length === 0
            ? <Typography variant="body2" color="text.secondary">The test reported no steps.</Typography>
            : <Box component="ol" sx={{listStyle: 'none', m: 0, p: 0}}>
                {steps.map((s, i) => <Step key={s.id} step={s} last={i === steps.length - 1} failed={view.failed}/>)}
                {!view.running && <Outcome view={view}/>}
            </Box>}
    </Section>;
};
