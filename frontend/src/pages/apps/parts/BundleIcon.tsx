import React, {useState} from 'react';
import {Avatar, Box} from '@mui/material';
import PlatformTypeIcon from '../../../components/PlatformTypeIcon';

interface BundleIconProps {
    // the url of the icon; without it, or when it doesn't load, the platform shows
    src?: string,
    platform?: number,
    size: number,
    radius: number,
}

const frame = {display: 'flex', alignItems: 'center', justifyContent: 'center', border: 1, borderColor: 'divider'} as const;

// BundleIcon is the app icon of a bundle, or the icon of its platform
export const BundleIcon: React.FC<BundleIconProps> = ({src, platform, size, radius}) => {
    const [failed, setFailed] = useState(false);
    if (src && !failed) {
        return <Avatar variant="rounded" src={src} alt="App icon" imgProps={{onError: () => setFailed(true)}}
            sx={{width: size, height: size, border: 1, borderColor: 'divider', borderRadius: radius}}/>;
    }
    return <Box sx={{...frame, width: size, height: size, borderRadius: radius}}>
        {platform !== undefined && <PlatformTypeIcon platformType={platform}/>}
    </Box>;
};
