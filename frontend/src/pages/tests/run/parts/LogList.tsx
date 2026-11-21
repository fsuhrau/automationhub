import React from 'react';
import {Box, Button, Link, Typography} from '@mui/material';
import {useTheme} from '@mui/material/styles';
import IProtocolEntryData from '../../../../types/protocol.entry';
import {formatRuntime, isError, isScreen, LogEntry, sourceName} from '../logModel';
import {LogView} from '../useLogView';
import {Message} from './Message';
import {Screenshot} from './Screenshot';

interface LogLineProps {
    entry: LogEntry,
    focused: boolean,
    // the index of the entry among the screenshots, for the viewer
    onScreen: (entry: IProtocolEntryData) => void,
}

const EntryBody: React.FC<LogLineProps> = ({entry, focused, onScreen}) =>
    isScreen(entry)
        ? <Box sx={{maxWidth: 240}}><Screenshot entry={entry} height={140} onClick={() => onScreen(entry)}/></Box>
        : <Message text={entry.message} error={isError(entry)} open={focused}/>;

const LogLine: React.FC<LogLineProps> = (props) => {
    const {entry, focused} = props;
    const theme = useTheme();
    return <Box id={`log-entry-${entry.id}`} sx={{
        display: 'grid', gridTemplateColumns: {xs: '72px 1fr', sm: '80px 140px 1fr'}, gap: 1.5, py: 0.5,
        borderLeft: 3, pl: 1, ml: -1.5,
        borderColor: isError(entry) ? (theme.vars ?? theme).palette.error.main : 'transparent',
        bgcolor: focused ? 'action.selected' : undefined,
    }}>
        <Typography variant="caption" color="text.secondary" sx={{fontFamily: 'monospace', pt: '2px'}}>{formatRuntime(entry.runtime)}</Typography>
        <Typography variant="caption" color="text.secondary" sx={{pt: '2px', display: {xs: 'none', sm: 'block'}, overflowWrap: 'anywhere'}}>
            {entry.phase && <Box component="span" sx={{color: 'secondary.main'}}>{entry.phase} · </Box>}
            {sourceName(entry.source)}
        </Typography>
        <EntryBody {...props}/>
    </Box>;
};

const Nothing: React.FC<{ onReset: () => void }> = ({onReset}) =>
    <Typography variant="body2" color="text.secondary" sx={{py: 2}}>
        Nothing matches.{' '}
        <Link component="button" variant="body2" onClick={onReset}>Show everything</Link>
    </Typography>;

interface LogListProps {
    logView: LogView,
    onScreen: (entry: IProtocolEntryData) => void,
}

// the entries of the log that pass the filters, a page at a time
export const LogList: React.FC<LogListProps> = ({logView, onScreen}) => {
    const {visible, limit, focus} = logView;
    if (visible.length === 0) return <Nothing onReset={logView.resetFilters}/>;
    return <Box role="log" aria-label="Protocol log" sx={{mt: 1}}>
        {visible.slice(0, limit).map(e => <LogLine key={e.id} entry={e} focused={focus === e.id} onScreen={onScreen}/>)}
        {visible.length > limit && <Button size="small" sx={{mt: 1}} onClick={logView.showMore}>
            Show {Math.min(500, visible.length - limit)} more of {visible.length - limit}
        </Button>}
    </Box>;
};
