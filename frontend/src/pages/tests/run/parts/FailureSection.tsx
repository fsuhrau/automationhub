import React from 'react';
import {Box, Button, Chip, Link, Stack, Typography} from '@mui/material';
import {useTheme} from '@mui/material/styles';
import {Section} from '../../../../components/common';
import {TestResultState} from '../../../../types/test.result.state.enum';
import {formatRuntime, LogEntry, sourceName} from '../logModel';
import {ProtocolView} from '../useProtocolView';
import {Message} from './Message';

const ErrorEntry: React.FC<{ entry: LogEntry, onShow: (e: LogEntry) => void }> = ({entry, onShow}) => {
    const theme = useTheme();
    const source = sourceName(entry.source);
    return <Box sx={{borderLeft: 3, borderColor: (theme.vars ?? theme).palette.error.main, pl: 1.5}}>
        <Stack direction="row" spacing={1} alignItems="center" useFlexGap={true} sx={{mb: 0.5}}>
            <Typography variant="caption" color="text.secondary" sx={{fontFamily: 'monospace'}}>{formatRuntime(entry.runtime)}</Typography>
            <Chip size="small" variant="outlined" label={entry.phase ? `${entry.phase} · ${source}` : source}/>
            <Link component="button" variant="caption" onClick={() => onShow(entry)} sx={{ml: 'auto'}}>Show in log</Link>
        </Stack>
        <Message text={entry.message} error={true}/>
    </Box>;
};

interface FailureSectionProps {
    view: ProtocolView,
    // shows an entry in the log tab
    onShow: (entry: LogEntry) => void,
    // shows the log tab with its errors only
    onAllErrors: () => void,
}

const ErrorList: React.FC<FailureSectionProps> = ({view, onShow}) => {
    const lastStep = view.summary.steps[view.summary.steps.length - 1];
    return <Stack spacing={1.5}>
        {lastStep && <Box>
            <Typography variant="caption" color="text.secondary">Last step, at {formatRuntime(lastStep.runtime)}</Typography>
            <Message text={lastStep.message}/>
        </Box>}
        {view.errors.slice(-3).map(e => <ErrorEntry key={`${e.id}`} entry={e} onShow={onShow}/>)}
    </Stack>;
};

// why a failed or unstable test went wrong: the last step and the last errors
export const FailureSection: React.FC<FailureSectionProps> = (props) => {
    const {view, onAllErrors} = props;
    const title = view.testResult === TestResultState.TestResultFailed ? 'Why it failed' : 'Why it is unstable';
    return <Section title={title} action={view.errors.length > 3 && <Button size="small" onClick={onAllErrors}>All {view.errors.length} errors</Button>}>
        {view.errors.length === 0
            ? <Typography variant="body2" color="text.secondary">The test reported no error message. The log shows what it did last.</Typography>
            : <ErrorList {...props}/>}
    </Section>;
};
