import React from 'react';
import {Box, Button, Stack, Typography} from '@mui/material';
import {CloudUploadRounded, CompareArrowsRounded} from '@mui/icons-material';
import {Link as RouterLink} from 'react-router-dom';
import PlatformTypeIcon from '../../../components/PlatformTypeIcon';
import {IAppData} from '../../../types/app';

interface BundlesHeaderProps {
    app?: IAppData,
    accept: string,
    // the link to compare bundles; none when there are fewer than two
    compareUrl?: string,
    onUpload: () => void,
}

const Title: React.FC<{ app?: IAppData }> = ({app}) =>
    <Stack direction="row" spacing={1.5} alignItems="center" sx={{minWidth: 0}}>
        {app && <Box sx={{display: 'flex', p: 1, borderRadius: 1, border: 1, borderColor: 'divider'}}>
            <PlatformTypeIcon platformType={app.platform}/>
        </Box>}
        <Box sx={{minWidth: 0}}>
            <Typography component="h1" variant="h5">Bundles</Typography>
            <Typography variant="body2" color="text.secondary" sx={{wordBreak: 'break-word'}}>
                Builds of {app?.name ?? 'this app'}
                {app?.identifier && <> (<Box component="span" sx={{fontFamily: 'monospace'}}>{app.identifier}</Box>)</>}
                {' '}to install on devices and run tests with.
            </Typography>
        </Box>
    </Stack>;

// BundlesHeader is the title of the list with the buttons to compare and upload
export const BundlesHeader: React.FC<BundlesHeaderProps> = ({app, accept, compareUrl, onUpload}) =>
    <Stack direction={{xs: 'column', md: 'row'}} spacing={2} useFlexGap={true} alignItems={{md: 'center'}} sx={{mb: 2}}>
        <Title app={app}/>
        <Stack direction="row" spacing={1} alignItems="center" sx={{ml: {md: 'auto'}}}>
            <Typography variant="caption" color="text.secondary" sx={{display: {xs: 'none', lg: 'block'}}}>
                or drop {accept} files on the page
            </Typography>
            {compareUrl && <Button variant="outlined" startIcon={<CompareArrowsRounded/>} component={RouterLink} to={compareUrl}>Compare</Button>}
            <Button variant="contained" startIcon={<CloudUploadRounded/>} onClick={onUpload}>Upload bundle</Button>
        </Stack>
    </Stack>;
