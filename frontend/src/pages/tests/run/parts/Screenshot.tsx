import React, {useState} from 'react';
import {Box, ButtonBase, Stack, Typography} from '@mui/material';
import {BrokenImageRounded} from '@mui/icons-material';
import IProtocolEntryData from '../../../../types/protocol.entry';
import {formatRuntime} from '../logModel';

const screenUrl = (e: IProtocolEntryData): string => `/api/data/${encodeURIComponent(e.data)}`;

const Missing: React.FC<{ height: number | string }> = ({height}) =>
    <Stack alignItems="center" justifyContent="center" spacing={0.5}
        sx={{height, width: '100%', bgcolor: 'action.hover', color: 'text.secondary', p: 1, textAlign: 'center'}}>
        <BrokenImageRounded/>
        <Typography variant="caption">The file is no longer on the master</Typography>
    </Stack>;

interface ScreenshotProps {
    entry: IProtocolEntryData,
    height: number | string,
    onClick?: () => void,
}

// Screenshot shows a captured screen, or says that its file is gone (old
// data is cleaned up on the master)
export const Screenshot: React.FC<ScreenshotProps> = ({entry, height, onClick}) => {
    const [broken, setBroken] = useState(false);
    const content = broken
        ? <Missing height={height}/>
        : <Box component="img" src={screenUrl(entry)} alt={entry.message || `Screenshot at ${formatRuntime(entry.runtime)}`}
                onError={() => setBroken(true)} loading="lazy"
                sx={{height, width: '100%', objectFit: 'contain', display: 'block', bgcolor: 'action.hover'}}/>;
    return onClick && !broken
        ? <ButtonBase onClick={onClick} sx={{display: 'block', width: '100%', borderRadius: 1, overflow: 'hidden'}}
                aria-label={`Open screenshot at ${formatRuntime(entry.runtime)}`}>{content}</ButtonBase>
        : <Box sx={{borderRadius: 1, overflow: 'hidden'}}>{content}</Box>;
};
