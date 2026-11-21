import React from 'react';
import {Alert} from '@mui/material';
import {getPlatformName, PlatformType} from '../../../types/platform.type.enum';

// PlatformNotice tells that apps of some platforms run without a bundle
export const PlatformNotice: React.FC<{ platform?: PlatformType }> = ({platform}) => {
    if (platform !== PlatformType.Editor && platform !== PlatformType.Web) return null;
    return <Alert severity="info" sx={{mb: 2}}>
        {getPlatformName(platform)} apps don't need a bundle: tests run in the {platform === PlatformType.Editor ? 'Unity editor' : 'browser'}.
    </Alert>;
};
