import React, {useMemo} from 'react';
import {ReplaySegment, segmentAt} from '../pages/tests/replayModel';
import {browserFrameIo, FrameIo} from '../pages/tests/replay/video/frameLoader';
import {MjpegSegment} from '../pages/tests/replay/video/MjpegSegment';
import {Mp4Segment} from '../pages/tests/replay/video/Mp4Segment';
import {VideoFrame, VideoMessage} from '../pages/tests/replay/video/VideoFrame';

// ReplayVideo shows a device's recording at a time of the replay clock. The
// clock is the page's: the video follows it, and is only moved when it is off
// by more than a little, so playing stays smooth.
interface ReplayVideoProps {
    segments: ReplaySegment[],
    // ms since the epoch
    position: number,
    playing: boolean,
    rate: number,
    label: string,
    // how MJPEG frames are fetched; tests inject a fake
    io?: FrameIo,
}

const ReplayVideo: React.FC<ReplayVideoProps> = ({segments, position, playing, rate, label, io = browserFrameIo}) => {
    const segment = segmentAt(segments, position);
    // keep the size of the frame between segments, so the page doesn't jump
    const sized = useMemo(
        () => segment ?? segments.find(s => s.start > position) ?? segments[segments.length - 1],
        [segment, segments, position]);
    return <VideoFrame width={sized?.width ?? 0} height={sized?.height ?? 0}>
        {!segment
            ? <VideoMessage center={true} text={segments.length === 0 ? 'This device has no recording.' : 'No video at this point of the run.'}/>
            : segment.codec === 'mp4'
                ? <Mp4Segment key={segment.id} segment={segment} position={position} playing={playing} rate={rate} label={label}/>
                : <MjpegSegment key={segment.id} segment={segment} position={position} playing={playing} label={label} io={io}/>}
    </VideoFrame>;
};

export default ReplayVideo;
