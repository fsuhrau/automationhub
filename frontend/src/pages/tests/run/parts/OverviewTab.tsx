import React from 'react';
import {Button, Stack, Typography} from '@mui/material';
import Grid from '@mui/material/Grid';
import {KeyValues, Section} from '../../../../components/common';
import IProtocolKeyData from '../../../../types/protocol.key';
import IProtocolEntryData from '../../../../types/protocol.entry';
import {formatRuntime, LogEntry} from '../logModel';
import {Comparison, TabKey} from '../protocolViewModel';
import {ProtocolView} from '../useProtocolView';
import {FailureSection} from './FailureSection';
import {MetricTiles} from './MetricTiles';
import {Screenshot} from './Screenshot';
import {StepsSection} from './StepsSection';

const keyRows = (keys: IProtocolKeyData[]): [string, React.ReactNode][] =>
    keys.map(k => [k.key, k.value || <Typography component="span" variant="body2" color="text.secondary">(empty)</Typography>]);

interface SideProps {
    view: ProtocolView,
    screens: LogEntry[],
    onTab: (tab: TabKey) => void,
    onViewer: (index: number) => void,
}

const LastScreen: React.FC<SideProps & { screen: IProtocolEntryData }> = ({view, screens, screen, onTab, onViewer}) =>
    <Section title={view.failed ? 'Screen at the end' : 'Last screenshot'} action={screens.length > 1 &&
        <Button size="small" onClick={() => onTab('screenshots')}>All {screens.length}</Button>}>
        <Screenshot entry={screen} height={420} onClick={() => onViewer(screens.length - 1)}/>
        <Typography variant="caption" color="text.secondary">at {formatRuntime(screen.runtime)}</Typography>
    </Section>;

// the right column: the last screen and the keys
const Side: React.FC<SideProps> = (props) => {
    const screen = props.screens[props.screens.length - 1];
    return <Stack spacing={2}>
        {screen && <LastScreen {...props} screen={screen}/>}
        {props.view.keys.length > 0 && <Section title="Keys"><KeyValues rows={keyRows(props.view.keys)} monospace={true}/></Section>}
    </Stack>;
};

interface OverviewTabProps extends SideProps {
    comparison: Comparison | null,
    onShow: (entry: LogEntry) => void,
    onAllErrors: () => void,
}

// why it failed, the steps, the last screen and the keys
export const OverviewTab: React.FC<OverviewTabProps> = (props) => {
    const {view, comparison, onTab, onShow, onAllErrors} = props;
    const side = props.screens.length > 0 || view.keys.length > 0;
    return <Grid container={true} spacing={2}>
        {view.metrics.length > 0 && <Grid size={12}>
            <Section title="Performance" action={<Button size="small" onClick={() => onTab('performance')}>Charts and checkpoints</Button>}>
                <MetricTiles protocol={view.protocol} metrics={view.metrics} against={comparison}/>
            </Section>
        </Grid>}
        <Grid size={{xs: 12, lg: side ? 8 : 12}}>
            <Stack spacing={2}>
                {view.failed && <FailureSection view={view} onShow={onShow} onAllErrors={onAllErrors}/>}
                <StepsSection view={view}/>
            </Stack>
        </Grid>
        {side && <Grid size={{xs: 12, lg: 4}}><Side {...props}/></Grid>}
    </Grid>;
};

