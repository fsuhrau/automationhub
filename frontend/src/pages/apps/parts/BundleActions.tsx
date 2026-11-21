import React from 'react';
import {Button, IconButton, Stack, Tooltip} from '@mui/material';
import {ArchiveOutlined, ArrowBackRounded, CompareArrowsRounded, DeleteOutlineRounded, DownloadRounded, RefreshRounded} from '@mui/icons-material';
import {Link as RouterLink} from 'react-router-dom';
import {AppRoutes} from '../../../lib/routes';
import {IAppBundleDetail} from '../../../types/app';
import {downloadUrl, isInstallable} from '../bundlesModel';

interface BundleActionsProps {
    detail: IAppBundleDetail,
    routes: AppRoutes,
    reading: boolean,
    onRead: () => void,
    onArchive: () => void,
    onDelete: () => void,
}

// BundleActions are the buttons on top of the details of a bundle
export const BundleActions: React.FC<BundleActionsProps> = ({detail, routes, reading, onRead, onArchive, onDelete}) => {
    const installable = isInstallable(detail);
    return <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap={true} sx={{ml: {md: 'auto'}}}>
        <Button startIcon={<ArrowBackRounded/>} component={RouterLink} to={routes.bundles()}>All bundles</Button>
        <Button variant="outlined" startIcon={<CompareArrowsRounded/>} component={RouterLink}
            to={`${routes.compareBundles()}?b=${detail.id}`}>Compare</Button>
        <Button variant="outlined" startIcon={<RefreshRounded/>} disabled={reading || !installable} onClick={onRead}>
            {reading ? 'Reading…' : 'Read again'}
        </Button>
        <Button variant="contained" startIcon={<DownloadRounded/>} disabled={!installable} href={downloadUrl(detail.appPath)} download={true}>Download</Button>
        {!detail.archivedAt && <Tooltip title="Archive: remove the file, keep the details">
            <IconButton aria-label="Archive bundle" onClick={onArchive}><ArchiveOutlined/></IconButton>
        </Tooltip>}
        <Tooltip title="Delete bundle">
            <IconButton color="error" aria-label="Delete bundle" onClick={onDelete}><DeleteOutlineRounded/></IconButton>
        </Tooltip>
    </Stack>;
};
