import React from 'react';
import {IBaselineCandidate} from '../../../../types/protocol.baseline';
import {ProtocolRef} from '../../../../types/test.run.refs';
import {LogEntry} from '../logModel';
import {ProtocolQuery} from '../useProtocolQuery';
import {ComparisonState} from '../useComparison';
import {LogView} from '../useLogView';
import {ProtocolView} from '../useProtocolView';
import {LogTab} from './LogTab';
import {OverviewTab} from './OverviewTab';
import {PerformanceTab} from './PerformanceTab';
import {ScreenshotsTab} from './ScreenshotsTab';

interface ProtocolBodyProps {
    refs: ProtocolRef,
    view: ProtocolView,
    query: ProtocolQuery,
    comparison: ComparisonState,
    logView: LogView,
    onViewer: (index: number) => void,
}

// the tab that is open
export const ProtocolBody: React.FC<ProtocolBodyProps> = ({refs, view, query, comparison, logView, onViewer}) => {
    const screens = view.logSummary.screens;
    const showInLog = (entry: LogEntry): void => {
        logView.showEntry(entry);
        query.setTab('log');
    };
    const onBaseline = (c: IBaselineCandidate | null): void => query.setBaseline(c, view.protocol.deviceId);
    switch (query.tab) {
        case 'log':
            return <LogTab view={view} logView={logView} onScreen={e => onViewer(screens.indexOf(e))}/>;
        case 'screenshots':
            return <ScreenshotsTab screens={screens} onViewer={onViewer}/>;
        case 'performance':
            return <PerformanceTab refs={refs} view={view} state={comparison} onBaseline={onBaseline} picked={!!query.vs}/>;
        default:
            return <OverviewTab view={view} screens={screens} comparison={comparison.comparison} onTab={query.setTab} onViewer={onViewer}
                onShow={showInLog} onAllErrors={() => {
                                    logView.setErrorsOnly(true);
                                    query.setTab('log');
                                }}/>;
    }
};
