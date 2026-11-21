import React from 'react';
import {Link as RouterLink} from 'react-router-dom';
import {Button, Checkbox, FormControlLabel, Link, Stack, Typography} from '@mui/material';
import {routes} from '../../../../lib/routes';
import {AppAction} from '../../../../types/device.control.api';
import {AppPicker} from '../useAppPicker';
import {ControlSession} from '../useControlSession';
import {AppSelect, BundleSelect} from './AppSelects';

const ACTIONS: { action: AppAction, label: string, variant: 'outlined' | 'contained', color?: 'warning' }[] = [
    {action: 'install', label: 'Install', variant: 'outlined'},
    {action: 'start', label: 'Start', variant: 'contained'},
    {action: 'stop', label: 'Stop', variant: 'outlined', color: 'warning'},
];

const NoBundles: React.FC<{ projectId: string, appId: number }> = ({projectId, appId}) => (
    <Typography variant="body2" color="text.secondary">
        This app has no bundles. <Link component={RouterLink} to={routes.project(projectId).app(appId).bundles()}>Bundles</Link>
    </Typography>
);

// install, start or stop a bundle on the device
const AppsTab: React.FC<{ session: ControlSession, picker: AppPicker }> = ({session, picker}) => {
    const hasLease = session.lease.token !== null;
    const unusable = !hasLease || picker.bundleId === '' || session.busy;
    return (
        <Stack spacing={2}>
            <Typography variant="body2" color="text.secondary">Install, start or stop a bundle on this device.</Typography>
            <AppSelect picker={picker}/>
            <BundleSelect picker={picker}/>
            {picker.appId !== '' && picker.bundles.length === 0 && <NoBundles projectId={session.target.projectId} appId={picker.appId}/>}
            <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap={true}>
                {ACTIONS.map(a => <Button key={a.action} variant={a.variant} color={a.color} disabled={unusable}
                    onClick={() => picker.act(a.action)}>{a.label}</Button>)}
            </Stack>
            <FormControlLabel label={<Typography variant="body2">Clear the app's data when stopping</Typography>}
                control={<Checkbox size="small" checked={picker.clearData} onChange={(_, v) => picker.setClearData(v)}/>}/>
            {!hasLease && <Typography variant="caption" color="text.secondary">Take control to use these.</Typography>}
        </Stack>
    );
};

export default AppsTab;
