import React, {useRef} from 'react';
import {Box, Button, FormControl, InputLabel, LinearProgress, ListItemText, MenuItem, Select, Stack, Typography} from '@mui/material';
import {CloudUploadRounded} from '@mui/icons-material';
import {formatDateTime} from '../../../../components/DateTime';
import {IAppBinaryData} from '../../../../types/app';
import {Accept} from '../../../apps/bundlesModel';
import {RunBundles} from './useRunBundles';

const label = (b: IAppBinaryData): string => b.version || b.name;

const BundleSelect: React.FC<{ bundles: RunBundles }> = ({bundles: state}) => {
    const {bundles, bundleId, choose} = state;
    return <FormControl size="small" fullWidth={true} disabled={bundles === null || bundles.length === 0}>
        <InputLabel id="run-bundle">Bundle</InputLabel>
        <Select labelId="run-bundle" label="Bundle" value={bundleId}
            onChange={e => choose(String(e.target.value) === '' ? '' : Number(e.target.value))}
            renderValue={id => {
                    const b = bundles?.find(x => x.id === id);
                    return b ? `${label(b)} (#${b.id})` : '';
                }}>
            {(bundles ?? []).map((b, i) => <MenuItem key={b.id} value={b.id}>
                <ListItemText primary={`${label(b)}${i === 0 ? ' (newest)' : ''}`}
                    secondary={`#${b.id} · uploaded ${formatDateTime(b.createdAt, 'YYYY/MM/DD HH:mm')}${b.tags ? ` · ${b.tags}` : ''}`}/>
            </MenuItem>)}
        </Select>
    </FormControl>;
};

const UploadButton: React.FC<{ accept: Accept, busy: boolean, onFile: (file: File) => void }> = ({accept, busy, onFile}) => {
    const input = useRef<HTMLInputElement>(null);
    return <>
        <Button variant="outlined" startIcon={<CloudUploadRounded/>} sx={{whiteSpace: 'nowrap', flexShrink: 0}}
            disabled={busy} onClick={() => input.current?.click()}>
            Upload
        </Button>
        <input ref={input} type="file" hidden={true} accept={accept.extensions.join(',')} aria-label="Bundle file"
            onChange={e => {
                   const file = e.target.files?.[0];
                   if (file) onFile(file);
                   e.target.value = '';
               }}/>
    </>;
};

const caption = (bundles: IAppBinaryData[] | null, accept: Accept): string => {
    if (bundles === null) return 'Loading bundles…';
    if (bundles.length === 0) return `No bundles yet; upload ${accept.label} to run the test.`;
    return 'The test installs this build on its devices first.';
};

const UploadProgress: React.FC<{ sent: number }> = ({sent}) =>
    <Box sx={{mt: 1}}>
        <Typography variant="caption">{sent < 1 ? `Uploading ${Math.round(sent * 100)} %` : 'Reading the bundle on the master…'}</Typography>
        <LinearProgress variant={sent < 1 ? 'determinate' : 'indeterminate'} value={sent * 100}/>
    </Box>;

// BundleField picks the bundle to install, or uploads a new one
export const BundleField: React.FC<{ bundles: RunBundles, accept: Accept }> = ({bundles, accept}) =>
    <Box>
        <Stack direction="row" spacing={1} alignItems="flex-start">
            <BundleSelect bundles={bundles}/>
            <UploadButton accept={accept} busy={bundles.progress !== null} onFile={bundles.upload}/>
        </Stack>
        <Typography variant="caption" color="text.secondary">{caption(bundles.bundles, accept)}</Typography>
        {bundles.progress !== null && <UploadProgress sent={bundles.progress}/>}
    </Box>;
