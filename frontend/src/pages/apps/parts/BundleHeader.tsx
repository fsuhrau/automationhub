import React from 'react';
import {Box, Chip, Stack, Typography} from '@mui/material';
import {ArchiveOutlined} from '@mui/icons-material';
import {IAppBundleDetail} from '../../../types/app';
import {badgesOf, bundleTitle, buildOf} from '../bundleDetailModel';
import {bundleVersion, parseTags} from '../bundlesModel';
import {BundleIcon} from './BundleIcon';

interface BundleHeaderProps {
    detail: IAppBundleDetail,
    iconUrl?: string,
    platform?: number,
}

const Badges: React.FC<{ detail: IAppBundleDetail }> = ({detail}) =>
    <Stack direction="row" spacing={0.5} flexWrap="wrap" useFlexGap={true} sx={{mt: 0.5}}>
        {badgesOf(detail, parseTags(detail.tags)).map(b =>
            <Chip key={b.label} size="small" label={b.label} color={b.warn ? 'warning' : 'default'} variant={b.outlined ? 'outlined' : 'filled'}/>)}
        {detail.archivedAt && <Chip size="small" variant="outlined" icon={<ArchiveOutlined/>} label="archived"/>}
    </Stack>;

// BundleHeader names a bundle: its icon, title, version and badges
export const BundleHeader: React.FC<BundleHeaderProps> = ({detail, iconUrl, platform}) => {
    const build = buildOf(detail);
    return <Stack direction="row" spacing={2} alignItems="center" sx={{minWidth: 0}}>
        <BundleIcon src={iconUrl} platform={platform} size={64} radius={3}/>
        <Box sx={{minWidth: 0}}>
            <Typography component="h1" variant="h5" sx={{wordBreak: 'break-word'}}>{bundleTitle(detail)}</Typography>
            <Typography variant="body2" color="text.secondary">
                Version {bundleVersion(detail) || '–'}{build && ` (build ${build})`} · #{detail.id}
            </Typography>
            <Badges detail={detail}/>
        </Box>
    </Stack>;
};
