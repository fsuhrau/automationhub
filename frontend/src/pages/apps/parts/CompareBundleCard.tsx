import React from 'react';
import {Box, Card, CardContent, Chip, Link, Stack, Tooltip, Typography} from '@mui/material';
import {ArchiveOutlined} from '@mui/icons-material';
import {Link as RouterLink} from 'react-router-dom';
import {formatDateTime} from '../../../components/DateTime';
import {IAppBundleDetail} from '../../../types/app';
import {ago, bundleLabel} from '../bundlesModel';
import {BundleIcon} from './BundleIcon';

interface CompareBundleCardProps {
    side: string,
    bundle: IAppBundleDetail,
    iconUrl?: string,
    to: string,
    platform?: number,
    now: number,
}

// CompareBundleCard is the head of one side of a comparison
export const CompareBundleCard: React.FC<CompareBundleCardProps> = ({side, bundle, iconUrl, to, platform, now}) =>
    <Card variant="outlined" sx={{flex: 1, minWidth: 0}}>
        <CardContent>
            <Typography variant="overline" color="text.secondary">{side}</Typography>
            <Stack direction="row" spacing={1.5} alignItems="center">
                <BundleIcon src={iconUrl} platform={platform} size={40} radius={1}/>
                <Box sx={{minWidth: 0}}>
                    <Link component={RouterLink} to={to} variant="subtitle1" sx={{fontWeight: 600, wordBreak: 'break-word'}}>
                        {bundleLabel(bundle)}
                    </Link>
                    <Typography variant="body2" color="text.secondary">
                        <Tooltip title={formatDateTime(bundle.createdAt)}><span>uploaded {ago(bundle.createdAt, now)}</span></Tooltip>
                        {' · '}{bundle.runCount} {bundle.runCount === 1 ? 'run' : 'runs'}
                    </Typography>
                    {bundle.archivedAt && <Chip size="small" variant="outlined" icon={<ArchiveOutlined/>} label="archived" sx={{mt: 0.5}}/>}
                </Box>
            </Stack>
            {!bundle.info && <Typography variant="caption" color="warning.main" sx={{display: 'block', mt: 1}}>
                Uploaded before the master read the details of a file; only the basics can be compared.
            </Typography>}
        </CardContent>
    </Card>;
