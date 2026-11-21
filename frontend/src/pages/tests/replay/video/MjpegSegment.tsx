import React from 'react';
import {CircularProgress} from '@mui/material';
import {FrameIo} from './frameLoader';
import {SegmentProps} from './Mp4Segment';
import {VideoMessage} from './VideoFrame';
import {useMjpegFrame} from './useMjpegFrame';

export const MjpegSegment: React.FC<SegmentProps & { io: FrameIo }> = ({segment, position, playing, label, io}) => {
    const {src, error} = useMjpegFrame(segment, io, {position, playing});
    if (error) return <VideoMessage text="This recording can't be played."/>;
    return src ? <img src={src} alt={label}/> : <CircularProgress size={28} sx={{color: 'grey.400'}}/>;
};
