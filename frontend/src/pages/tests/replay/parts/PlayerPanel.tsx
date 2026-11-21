import React from 'react';
import {Alert, Card, CardContent} from '@mui/material';
import ReplayVideo from '../../../../components/ReplayVideo';
import {nextMarkAfter} from '../eventView';
import {ReplayClock} from '../useReplayClock';
import {ReplayDevice} from '../useReplayDevice';
import {DeviceChips} from './DeviceChips';
import {ReplayControls} from './ReplayControls';
import {TimelineSlider} from './TimelineSlider';

interface PlayerPanelProps {
    device: ReplayDevice,
    clock: ReplayClock,
    start: number,
    end: number,
}

export const PlayerPanel: React.FC<PlayerPanelProps> = ({device, clock, start, end}) => {
    const {deviceIds, deviceId, deviceSegments, events, nameOf, pick} = device;
    const {position, playing, rate, seek} = clock;
    return <Card variant="outlined">
        <CardContent>
            {deviceIds.length > 1 && <DeviceChips ids={deviceIds} selected={deviceId} nameOf={nameOf}
                onPick={id => pick(id, position)}/>}
            <ReplayVideo segments={deviceSegments} position={position} playing={playing} rate={rate}
                label={`Screen of ${deviceId !== null ? nameOf(deviceId) : 'the device'}`}/>
            <TimelineSlider events={events} start={start} end={end} position={position} onSeek={seek}/>
            <ReplayControls clock={clock} start={start} end={end} nextMark={nextMarkAfter(events, position)}/>
            {deviceSegments.length === 0 && <Alert severity="info" variant="outlined" sx={{mt: 1}}>
                This device has no recording. Turn on "Record the session" in the test settings to record the next runs.
            </Alert>}
        </CardContent>
    </Card>;
};
