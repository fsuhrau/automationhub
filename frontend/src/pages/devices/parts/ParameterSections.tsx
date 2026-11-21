import React from 'react';
import {Button, Link, Typography} from '@mui/material';
import {Link as RouterLink} from 'react-router-dom';
import IParameter from '../../../types/device.parameter';
import {KeyValues, Section} from '../../../components/common';

const Note: React.FC<{ children: React.ReactNode }> = ({children}) =>
    <Typography variant="body2" color="text.secondary">{children}</Typography>;

const rowsOf = (parameters: IParameter[]): [string, string][] => parameters.map(p => [p.key, p.value]);

export const ReportedSection: React.FC<{ parameters: IParameter[] }> = ({parameters}) =>
    <Section title="Reported by the device">
        {parameters.length === 0
            ? <Note>The device reported no further details.</Note>
            : <KeyValues rows={rowsOf(parameters)}/>}
    </Section>;

const NoCustom: React.FC<{ editUrl: string }> = ({editUrl}) =>
    <Note>
        None. Custom parameters are passed to tests on this device; add them under <Link component={RouterLink} to={editUrl}>Edit</Link>.
    </Note>;

export const CustomSection: React.FC<{ parameters: IParameter[], editUrl: string }> = ({parameters, editUrl}) =>
    <Section title="Custom parameters" action={<Button size="small" component={RouterLink} to={editUrl}>Edit</Button>}>
        {parameters.length === 0
            ? <NoCustom editUrl={editUrl}/>
            : <KeyValues rows={rowsOf(parameters)}/>}
    </Section>;
