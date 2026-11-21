import React, {useMemo, useRef, useState} from 'react';
import {Box, Card, CardContent, Chip, Stack, Tooltip, Typography} from '@mui/material';
import {eventIndexAt, ReplayEvent} from '../../replayModel';
import {EventFilter, FILTERS, shownBy} from '../eventView';
import {useFollowScroll} from '../useFollowScroll';
import {EventRow} from './EventRow';

interface Toggle {
    label: string,
    on: boolean,
    onClick: () => void,
}

const ToggleChip: React.FC<Toggle> = ({label, on, onClick}) =>
    <Chip size="small" label={label} color={on ? 'primary' : 'default'} variant={on ? 'filled' : 'outlined'} onClick={onClick}/>;

const FilterChips: React.FC<{ filter: EventFilter, onChange: (f: EventFilter) => void }> = ({filter, onChange}) =>
    <Stack direction="row" spacing={1} role="group" aria-label="Show">
        {FILTERS.map(([f, label]) => <ToggleChip key={f} label={label} on={filter === f} onClick={() => onChange(f)}/>)}
    </Stack>;

interface EventsPanelProps {
    events: ReplayEvent[],
    start: number,
    position: number,
    playing: boolean,
    onSeek: (time: number) => void,
}

export const EventsPanel: React.FC<EventsPanelProps> = ({events, start, position, playing, onSeek}) => {
    const [filter, setFilter] = useState<EventFilter>('all');
    const [follow, setFollow] = useState(true);
    const visible = useMemo(() => events.filter(shownBy(filter)), [events, filter]);
    const activeIndex = eventIndexAt(visible, position);
    const list = useRef<HTMLDivElement>(null);
    const activeRow = useRef<HTMLDivElement>(null);
    useFollowScroll(list, activeRow, {follow, activeKey: activeIndex >= 0 ? visible[activeIndex].key : null, playing});
    const stopFollowing = (): void => setFollow(false);
    return <Card variant="outlined">
        <CardContent>
            <Stack direction="row" alignItems="center" spacing={1} flexWrap="wrap" useFlexGap={true} sx={{mb: 1}}>
                <Typography component="h2" variant="subtitle1" sx={{fontWeight: 600, mr: 1}}>What happened</Typography>
                <FilterChips filter={filter} onChange={setFilter}/>
                <Tooltip title="Keep the current line in view while the video plays">
                    <span style={{marginLeft: 'auto'}}>
                        <ToggleChip label="Follow the video" on={follow} onClick={() => setFollow(f => !f)}/>
                    </span>
                </Tooltip>
            </Stack>
            {visible.length === 0
                ? <Typography variant="body2" color="text.secondary">Nothing to show.</Typography>
                : <Box ref={list} role="list" aria-label="What happened in the run" onWheel={stopFollowing} onTouchMove={stopFollowing}
                        sx={{maxHeight: {xs: 480, lg: '75vh'}, overflowY: 'auto', position: 'relative', px: 1}}>
                    {visible.map((e, i) => <EventRow key={e.key} event={e} offset={e.time - start} active={i === activeIndex}
                        future={e.time > position} onSeek={onSeek} rowRef={i === activeIndex ? activeRow : undefined}/>)}
                </Box>}
        </CardContent>
    </Card>;
};
