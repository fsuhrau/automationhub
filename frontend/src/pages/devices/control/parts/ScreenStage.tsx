import React, {CSSProperties, RefObject} from 'react';
import {Box} from '@mui/material';
import {ControlSession} from '../useControlSession';
import {ScreenState} from '../useScreen';
import LeaseOverlay from './LeaseOverlay';

interface ScreenStageProps {
    session: ControlSession,
    screen: ScreenState,
    // the element picked in the scene graph, in percent of the screen
    highlight: Record<string, string> | null,
    picking: boolean,
    focused: boolean,
    onFocusChange: (focused: boolean) => void,
}

const Highlight: React.FC<{ style: Record<string, string> }> = ({style}) => (
    <Box data-testid="highlight" sx={{
        position: 'absolute', pointerEvents: 'none', border: '2px solid', borderColor: 'error.main',
        backgroundColor: 'rgba(255,0,0,0.15)', ...style,
    }}/>
);

const imageStyle = (picking: boolean, hasLease: boolean): CSSProperties => ({
    maxHeight: '72vh', maxWidth: '100%', display: 'block', userSelect: 'none', touchAction: 'none',
    cursor: picking ? 'help' : hasLease ? 'crosshair' : 'default',
});

interface ScreenImageProps {
    video: boolean,
    canvasRef: RefObject<HTMLCanvasElement | null>,
    src: string,
    style: CSSProperties,
    handlers: object,
}

const ScreenImage: React.FC<ScreenImageProps> = ({video, canvasRef, src, style, handlers}) =>
    video
        ? <canvas ref={canvasRef} width={360} height={640} {...handlers} style={{...style, backgroundColor: '#000'}}/>
        : <img src={src} alt="device screen" draggable={false} {...handlers} style={style}/>;

// the screen: video or screenshot with the picked element and, without
// control, the overlay. Focus makes the keyboard type on the device.
const ScreenStage: React.FC<ScreenStageProps> = ({session, screen, highlight, picking, focused, onFocusChange}) => {
    const hasLease = session.lease.token !== null;
    const handlers = hasLease || picking ? screen.pointer : {};
    return (
        <Box tabIndex={hasLease ? 0 : -1} onKeyDown={screen.typing.onScreenKey} aria-label="Device screen"
            onFocus={() => onFocusChange(true)}
            onBlur={() => {
                 onFocusChange(false);
                 screen.typing.flush();
             }}
            sx={{
                 // clip the video, screenshot and overlay to the rounded corners
                 position: 'relative', display: 'inline-block', lineHeight: 0, borderRadius: 1, outline: 'none', overflow: 'hidden',
                 boxShadow: focused ? (t => `0 0 0 2px ${(t.vars ?? t).palette.primary.main}`) : undefined,
             }}>
            <ScreenImage video={screen.modes.mode === 'video'} canvasRef={screen.video.canvasRef} src={screen.source.src}
                style={imageStyle(picking, hasLease)} handlers={handlers}/>
            {highlight && <Highlight style={highlight}/>}
            {!hasLease && !picking && <LeaseOverlay session={session}/>}
        </Box>
    );
};

export default ScreenStage;
