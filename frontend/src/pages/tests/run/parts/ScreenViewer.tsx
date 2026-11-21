import React from 'react';
import {Box, Dialog, DialogContent, IconButton, Stack, Typography} from '@mui/material';
import {ChevronLeftRounded, ChevronRightRounded, CloseRounded} from '@mui/icons-material';
import IProtocolEntryData from '../../../../types/protocol.entry';
import {formatRuntime} from '../logModel';
import {Screenshot} from './Screenshot';

interface ScreenViewerProps {
    screens: IProtocolEntryData[],
    // the index of the shown screenshot, null while closed
    index: number | null,
    onIndex: (index: number | null) => void,
}

const ViewerBar: React.FC<ScreenViewerProps & { index: number }> = ({screens, index, onIndex}) =>
    <Stack direction="row" alignItems="center" spacing={1} sx={{mb: 1, px: 1}}>
        <Typography variant="body2" color="text.secondary">
            {index + 1} of {screens.length} · at {formatRuntime(screens[index].runtime)}
        </Typography>
        <Box sx={{ml: 'auto'}}/>
        <IconButton size="small" aria-label="Previous screenshot" disabled={index === 0}
            onClick={() => onIndex(index - 1)}><ChevronLeftRounded/></IconButton>
        <IconButton size="small" aria-label="Next screenshot" disabled={index === screens.length - 1}
            onClick={() => onIndex(index + 1)}><ChevronRightRounded/></IconButton>
        <IconButton size="small" aria-label="Close" onClick={() => onIndex(null)}><CloseRounded/></IconButton>
    </Stack>;

// the screenshots one by one in a dialog
export const ScreenViewer: React.FC<ScreenViewerProps> = (props) => {
    const {screens, index, onIndex} = props;
    return <Dialog open={index !== null} onClose={() => onIndex(null)} maxWidth="lg" fullWidth={true} aria-label="Screenshot">
        {index !== null && screens[index] && <DialogContent sx={{p: 1}}>
            <ViewerBar {...props} index={index}/>
            <Screenshot key={screens[index].id} entry={screens[index]} height="80vh"/>
        </DialogContent>}
    </Dialog>;
};
