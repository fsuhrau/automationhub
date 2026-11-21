import React from 'react';
import {Box, FormControlLabel, RadioGroup, Switch, Typography} from '@mui/material';
import {Section} from '../../../../components/common';
import IDeviceData from '../../../../types/device';
import {TestExecutionType} from '../../../../types/test.execution.type.enum';
import {FormErrors, TestForm} from '../../testEditModel';
import {Choice} from './Choice';
import {DevicePicker} from './DevicePicker';

interface WhereItRunsSectionProps {
    form: TestForm,
    set: (change: Partial<TestForm>) => void,
    errors: FormErrors,
    devices: IDeviceData[] | null,
    appName?: string,
}

const ExecutionChoice: React.FC<Pick<WhereItRunsSectionProps, 'form' | 'set'>> = ({form, set}) =>
    <>
        <Typography variant="body2" sx={{fontWeight: 600, mt: 1}}>With several devices</Typography>
        <RadioGroup value={String(form.executionType)} onChange={e => set({executionType: Number(e.target.value) as TestExecutionType})}>
            <Choice value={String(TestExecutionType.Concurrent)} label="Spread the tests"
                hint="Each test runs on the next free device. Faster results."/>
            <Choice value={String(TestExecutionType.Simultaneously)} label="Every test on every device"
                hint="Each test runs on all devices. Shows differences between devices."/>
        </RadioGroup>
    </>;

const RecordingSwitch: React.FC<Pick<WhereItRunsSectionProps, 'form' | 'set'>> = ({form, set}) =>
    <>
        <Typography variant="body2" sx={{fontWeight: 600, mt: 1}}>Recording</Typography>
        <FormControlLabel sx={{alignItems: 'flex-start', '& .MuiSwitch-root': {mt: -0.5}}}
            control={<Switch size="small" checked={form.recordSession} onChange={e => set({recordSession: e.target.checked})}/>}
            label={<Box>
                <Typography variant="body2" sx={{fontWeight: 600}}>Record the session</Typography>
                <Typography variant="caption" color="text.secondary">
                    Records the screen of each device. After the run, its replay shows the video next to the log and the test steps.
                </Typography>
            </Box>}/>
    </>;

// WhereItRunsSection picks the devices, how tests spread over them, and
// whether the sessions are recorded
export const WhereItRunsSection: React.FC<WhereItRunsSectionProps> = ({form, set, errors, devices, appName}) =>
    <Section title="Where it runs">
        <RadioGroup value={form.allDevices ? 'all' : 'picked'} onChange={e => set({allDevices: e.target.value === 'all'})}>
            <Choice value="all" label="All devices" hint="Every device in the device pool."/>
            <Choice value="picked" label="Picked devices" hint="Only the devices picked below."/>
        </RadioGroup>
        {!form.allDevices && <DevicePicker devices={devices} picked={form.devices} onChange={ids => set({devices: ids})}
            appName={appName} error={errors.devices}/>}
        <ExecutionChoice form={form} set={set}/>
        <RecordingSwitch form={form} set={set}/>
    </Section>;
