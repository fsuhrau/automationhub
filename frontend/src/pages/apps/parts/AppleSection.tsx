import React from 'react';
import {Stack, Typography} from '@mui/material';
import {Section} from '../../../components/common';
import {AppleInfo, BundleInfo} from '../../../types/bundle.info';
import {appleFacts} from '../bundleFactsModel';
import {usageFacts} from '../signingFactsModel';
import {Facts} from './Facts';

export const AppleSection: React.FC<{ apple: AppleInfo, info: BundleInfo, now: number }> = ({apple, info, now}) =>
    <Section title={apple.platform === 'macosx' ? 'macOS' : 'iOS'} description="From the Info.plist" fullHeight={true}>
        <Facts facts={appleFacts(apple, info)} now={now}/>
    </Section>;

export const UsageSection: React.FC<{ apple: AppleInfo, now: number }> = ({apple, now}) => {
    const usage = apple.usageDescriptions ?? [];
    return <Section title={`Permissions (${usage.length})`} description="What the app tells the player when it asks for access" fullHeight={true}>
        {usage.length === 0
            ? <Typography variant="body2" color="text.secondary">The app asks for no protected access.</Typography>
            : <Stack spacing={0}><Facts facts={usageFacts(usage)} now={now}/></Stack>}
    </Section>;
};
