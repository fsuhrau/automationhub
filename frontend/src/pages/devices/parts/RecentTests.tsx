import React from 'react';
import {Table, TableBody, Typography} from '@mui/material';
import ITestProtocolData from '../../../types/test.protocol';
import {Section} from '../../../components/common';
import {useLoader} from '../../../hooks/useLoader';
import {useDeviceService} from '../../../services/device.service';
import {RecentTestRow} from './RecentTestRow';

interface RecentTestsProps {
    projectId: string,
    deviceId: number,
}

const Note: React.FC<{ text: string }> = ({text}) =>
    <Typography variant="body2" color="text.secondary">{text}</Typography>;

const Results: React.FC<{ projectId: string, protocols: ITestProtocolData[] }> = ({projectId, protocols}) =>
    <Table size="small" aria-label="Recent tests on this device">
        <TableBody>
            {protocols.map(p => <RecentTestRow key={p.id as number} projectId={projectId} protocol={p}/>)}
        </TableBody>
    </Table>;

// RecentTests lists what ran on the device lately
export const RecentTests: React.FC<RecentTestsProps> = ({projectId, deviceId}) => {
    const service = useDeviceService();
    const {data, error} = useLoader(() => service.protocols(projectId, deviceId, 10), [service, projectId, deviceId]);
    return <Section title="Recent tests">
        {error ? <Note text="The test history couldn't be loaded."/> : null}
        {!error && data === null ? <Note text="Loading…"/> : null}
        {!error && data?.length === 0 ? <Note text="No test of this project ran on this device yet."/> : null}
        {!error && data && data.length > 0 ? <Results projectId={projectId} protocols={data}/> : null}
    </Section>;
};
