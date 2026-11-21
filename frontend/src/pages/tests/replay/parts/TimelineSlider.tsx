import React, {useMemo} from 'react';
import {Box, Slider} from '@mui/material';
import {formatOffset, ReplayEvent} from '../../replayModel';
import {isMarker, markerColor} from '../eventView';

const MAX_MARKS = 600;

interface TimelineProps {
    events: ReplayEvent[],
    start: number,
    end: number,
}

// Markers on the timeline: where tests ended (by result) and errors happened
const Markers: React.FC<TimelineProps> = ({events, start, end}) => {
    const span = Math.max(1, end - start);
    const marks = useMemo(() => events.filter(isMarker).slice(0, MAX_MARKS), [events]);
    return <Box aria-hidden="true" data-testid="markers" sx={{position: 'relative', height: 8, mx: '10px'}}>
        {marks.map(e => <Box key={e.key} sx={{
            position: 'absolute', left: `${((e.time - start) / span) * 100}%`, top: 0, width: 2, height: 8,
            bgcolor: markerColor(e), borderRadius: 1, transform: 'translateX(-1px)',
        }}/>)}
    </Box>;
};

interface TimelineSliderProps extends TimelineProps {
    position: number,
    onSeek: (time: number) => void,
}

export const TimelineSlider: React.FC<TimelineSliderProps> = ({events, start, end, position, onSeek}) =>
    <Box sx={{mt: 1.5}}>
        <Markers events={events} start={start} end={end}/>
        <Slider size="small" min={0} max={Math.max(1, end - start)} value={position - start}
            onChange={(_, v) => onSeek(start + (v as number))}
            aria-label="Position in the run"
            valueLabelDisplay="auto" valueLabelFormat={v => formatOffset(v)}
            sx={{mt: -0.5}}/>
    </Box>;
