import React, {useState} from 'react';
import {CircularProgress} from '@mui/material';
import {ReplaySegment} from '../../replayModel';
import {VideoMessage} from './VideoFrame';
import {useVideoSync} from './useVideoSync';

export interface SegmentProps {
    segment: ReplaySegment,
    // ms since the epoch
    position: number,
    playing: boolean,
    label: string,
}

export const Mp4Segment: React.FC<SegmentProps & { rate: number }> = ({segment, position, playing, rate, label}) => {
    const [error, setError] = useState(false);
    const [loading, setLoading] = useState(true);
    const video = useVideoSync({position, segmentStart: segment.start, playing, rate, ready: !loading});

    if (error) return <VideoMessage text="This recording can't be played."/>;
    return <>
        <video ref={video} src={segment.url} muted={true} playsInline={true} preload="auto" aria-label={label}
            onLoadedMetadata={() => setLoading(false)} onError={() => setError(true)}/>
        {loading && <CircularProgress size={28} sx={{position: 'absolute', color: 'grey.400'}}/>}
    </>;
};
