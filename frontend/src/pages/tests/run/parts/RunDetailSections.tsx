import React from 'react';
import {Box, Link, Typography} from '@mui/material';
import {Link as RouterLink} from 'react-router-dom';
import {formatDateTime} from '../../../../components/DateTime';
import {KeyValues, Section} from '../../../../components/common';
import {routes} from '../../../../lib/routes';
import ITestRunData from '../../../../types/test.run';
import ITesRunDeviceStatusData from '../../../../types/test.run.device.status';
import {parseParameters} from '../../testRunModel';
import {deviceLabel} from '../runStatusModel';

const Muted: React.FC<{ children: React.ReactNode }> = ({children}) =>
    <Typography variant="body2" color="text.secondary">{children}</Typography>;

const startupColor = (status: ITesRunDeviceStatusData): string => {
    const avg = status.histAvgStartupTime;
    if (avg > 0 && status.startupTime > avg * 1.2) return 'warning.main';
    if (avg > 0 && status.startupTime < avg * 0.8) return 'success.main';
    return 'text.primary';
};

const DeviceStartup: React.FC<{ projectId: string, status: ITesRunDeviceStatusData }> = ({projectId, status}) => {
    const avg = status.histAvgStartupTime;
    return <Box sx={{py: 0.5}}>
        <Link component={RouterLink} to={routes.project(projectId).device(status.deviceId).show()}
            variant="body2" underline="hover" sx={{fontWeight: 600}}>
            {deviceLabel(status.device) || `Device ${status.deviceId}`}
        </Link>
        <Muted>
            App started in{' '}
            <Box component="span" sx={{color: startupColor(status), fontWeight: 600}}>{(status.startupTime / 1000).toFixed(1)} s</Box>
            {avg > 0 && <> (usually {(avg / 1000).toFixed(1)} s)</>}
        </Muted>
    </Box>;
};

// the devices that started the app and how long that took
export const DevicesSection: React.FC<{ run: ITestRunData, running: boolean, projectId: string }> = ({run, running, projectId}) =>
    <Section title="Devices">
        {run.deviceStatus.length === 0
            ? <Muted>{running ? 'No device has started yet.' : 'No device started the app.'}</Muted>
            : run.deviceStatus.map(d => <DeviceStartup key={d.id ?? d.deviceId} projectId={projectId} status={d}/>)}
    </Section>;

const parameterRows = (run: ITestRunData): [string, React.ReactNode][] => [
    ...(run.startUrl ? [['Start URL', run.startUrl] as [string, React.ReactNode]] : []),
    ...parseParameters(run.parameter).map((p): [string, React.ReactNode] => [p.name, p.value || <Muted>(empty)</Muted>]),
];

export const ParametersSection: React.FC<{ run: ITestRunData }> = ({run}) => {
    const rows = parameterRows(run);
    return <Section title="Parameters">
        {rows.length === 0 ? <Muted>The run had no parameters.</Muted> : <KeyValues rows={rows} monospace={true}/>}
    </Section>;
};

const bundleRows = (binary: NonNullable<ITestRunData['appBinary']>): [string, React.ReactNode][] => [
    ['Name', binary.name],
    ['Version', binary.version],
    ['Identifier', binary.identifier],
    ['Uploaded', formatDateTime(binary.createdAt)],
    ...(binary.additional ? [['Addons', binary.additional] as [string, React.ReactNode]] : []),
    ['Hash', binary.hash],
];

export const BundleSection: React.FC<{ run: ITestRunData, bundlesUrl: string }> = ({run, bundlesUrl}) => {
    const binary = run.appBinary;
    return <Section title="App bundle" action={binary && <Link component={RouterLink} to={bundlesUrl} variant="body2">All bundles</Link>}>
        {binary
            ? <KeyValues rows={bundleRows(binary)} monospace={true}/>
            : <Muted>No bundle was installed for this run; it used the app already on the device or editor.</Muted>}
    </Section>;
};
