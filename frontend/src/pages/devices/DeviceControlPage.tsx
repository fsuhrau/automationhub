import React, {useMemo, useState} from 'react';
import {Alert, Box} from '@mui/material';
import {VideoSupport} from '../../types/device.control.video';
import {PlatformType} from '../../types/platform.type.enum';
import {realVideoSupport} from '../../services/device.control.video';
import ControlHeader from './control/parts/ControlHeader';
import ControlTabs, {ControlTab} from './control/parts/ControlTabs';
import ScreenCard from './control/parts/ScreenCard';
import {useControlSession} from './control/useControlSession';
import {useSceneGraph} from './control/useSceneGraph';
import {useScreen} from './control/useScreen';

interface DeviceControlPageProps {
    // what the browser can decode and how to play it; tests fake it
    support?: VideoSupport,
}

// The control page of a device: take control of it, watch and use its
// screen, inspect the running app, read the system log and manage the app.
const DeviceControlPage: React.FC<DeviceControlPageProps> = ({support}) => {
    const video = useMemo(() => support ?? realVideoSupport(), [support]);
    const session = useControlSession();
    const scene = useSceneGraph(session.service, session.run);
    const [tab, setTab] = useState<ControlTab>('app');
    const onPick = scene.picking ? (el: HTMLElement, at: { x: number, y: number }) => {
        scene.pickAt(el, at);
        setTab('inspect');
    } : null;
    const screen = useScreen(session, {support: video, onPick});
    const suggestVideo = session.lease.token !== null && screen.modes.mode !== 'video'
        && session.device?.platformType === PlatformType.Android && video.decodesVideo;

    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}, mt: 2}}>
            <ControlHeader session={session}/>
            <Box sx={{display: 'flex', gap: 2, flexWrap: 'wrap', alignItems: 'flex-start'}}>
                <ScreenCard session={session} screen={screen} highlight={scene.highlight} picking={scene.picking}/>
                <ControlTabs session={session} scene={scene} tab={tab} onTab={setTab}/>
            </Box>
            {suggestVideo && <Alert severity="info" sx={{mt: 2}}>Video mode is faster and sends touches and keys live on Android.</Alert>}
        </Box>
    );
};

export default DeviceControlPage;
