import React, {useState} from 'react';
import {Box, Card, CardContent, Typography} from '@mui/material';
import {KeyboardRounded} from '@mui/icons-material';
import {ControlSession} from '../useControlSession';
import {ScreenState} from '../useScreen';
import {NavKeys, VolumeKeys} from './KeyPad';
import ScreenStage from './ScreenStage';
import ScreenToolbar from './ScreenToolbar';
import TypeField from './TypeField';

interface ScreenCardProps {
    session: ControlSession,
    screen: ScreenState,
    highlight: Record<string, string> | null,
    picking: boolean,
}

const KeyboardHint: React.FC<{ focused: boolean }> = ({focused}) => (
    <Typography variant="caption" color="text.secondary" component="div" sx={{mt: 0.5, display: 'flex', gap: 0.5, alignItems: 'center'}}>
        <KeyboardRounded fontSize="inherit"/>
        {focused ? 'Your keyboard types on the device; Esc is Back.' : 'Click the screen to type on it with your keyboard.'}
    </Typography>
);

// the screen with what can be done on it
const ScreenCard: React.FC<ScreenCardProps> = ({session, screen, highlight, picking}) => {
    const [focused, setFocused] = useState(false);
    const noLease = session.lease.token === null;
    return (
        <Card variant="outlined" sx={{flex: '0 1 auto', maxWidth: '100%'}}>
            <CardContent sx={{p: 1.5, '&:last-child': {pb: 1.5}}}>
                <ScreenToolbar screen={screen} busy={session.busy}/>
                <Box sx={{display: 'flex', alignItems: 'flex-start', gap: 1, maxWidth: '100%'}}>
                    <Box sx={{display: 'flex', flexDirection: 'column', alignItems: 'center', minWidth: 0}}>
                        <ScreenStage session={session} screen={screen} highlight={highlight} picking={picking}
                            focused={focused} onFocusChange={setFocused}/>
                        <NavKeys disabled={noLease} onKey={screen.input.sendKey}/>
                    </Box>
                    <VolumeKeys disabled={noLease} onKey={screen.input.sendKey}/>
                </Box>
                <TypeField input={screen.input} disabled={noLease}/>
                {!noLease && <KeyboardHint focused={focused}/>}
            </CardContent>
        </Card>
    );
};

export default ScreenCard;
