import React, {useEffect} from 'react';
import {Card, CardContent, Typography} from '@mui/material';
import IProtocolEntryData from '../../../../types/protocol.entry';
import {logAsText} from '../logModel';
import {downloadText} from '../downloadText';
import {LogView} from '../useLogView';
import {ProtocolView} from '../useProtocolView';
import {LogList} from './LogList';
import {LogToolbar} from './LogToolbar';

interface LogTabProps {
    view: ProtocolView,
    logView: LogView,
    onScreen: (entry: IProtocolEntryData) => void,
}

const Count: React.FC<{ view: ProtocolView, logView: LogView }> = ({view, logView}) => {
    const shown = logView.visible.length;
    const total = view.log.length;
    return <Typography variant="caption" color="text.secondary">
        {shown === total ? `${total} entries` : `${shown} of ${total} entries`}
        {view.phases.length > 0 ? ' · times since the test started, setup before 0:00' : ' · times since the test started'}
    </Typography>;
};

// scrolls the entry "show in log" jumped to into view
const useScrollToFocus = (focus: number | null): void => {
    useEffect(() => {
        if (focus !== null) document.getElementById(`log-entry-${focus}`)?.scrollIntoView({block: 'center'});
    }, [focus]);
};

// the whole log of the test and its runner, to filter, read and download
export const LogTab: React.FC<LogTabProps> = ({view, logView, onScreen}) => {
    const {protocol, log, name} = view;
    useScrollToFocus(logView.focus);
    const download = (): void => downloadText(`protocol-${protocol.id}-${name}.txt`, logAsText(protocol.testName, log));
    return <Card variant="outlined">
        <CardContent>
            <LogToolbar view={view} logView={logView} onDownload={download}/>
            <Count view={view} logView={logView}/>
            <LogList logView={logView} onScreen={onScreen}/>
        </CardContent>
    </Card>;
};
