import React from 'react';
import {Alert, Link, Stack, Typography} from '@mui/material';
import {Link as RouterLink} from 'react-router-dom';
import {Section} from '../../../../components/common';
import {ProjectRoutes} from '../../../../lib/routes';
import ITestConfigData from '../../../../types/test.config';
import {TestExecutionType} from '../../../../types/test.execution.type.enum';

const DeviceLinks: React.FC<{ config: ITestConfigData, projectRoutes: ProjectRoutes }> = ({config, projectRoutes}) => {
    if (config.allDevices) return <Typography variant="body2" color="text.secondary">On all devices.</Typography>;
    if ((config.devices ?? []).length === 0) {
        return <Alert severity="warning">No devices are picked, so the test can't run. Pick devices under Edit.</Alert>;
    }
    return <Stack spacing={0.5}>
        {config.devices.map((d, i) => <Link key={`${d.deviceId}_${i}`} component={RouterLink} underline="hover"
            to={projectRoutes.device(d.deviceId).show()} variant="body2">
            {d.device ? (d.device.alias || d.device.name) : `Device ${d.deviceId}`}
        </Link>)}
    </Stack>;
};

// WhereItRunsCard says how the test spreads over devices and which they are
export const WhereItRunsCard: React.FC<{ config: ITestConfigData, projectRoutes: ProjectRoutes }> = ({config, projectRoutes}) =>
    <Section title="Where it runs">
        <Typography variant="body2" sx={{mb: 1}}>
            {config.executionType === TestExecutionType.Simultaneously
                ? 'Every test runs on every device, for results across devices.'
                : 'Each test runs on the next free device, for faster results.'}
        </Typography>
        <DeviceLinks config={config} projectRoutes={projectRoutes}/>
    </Section>;
