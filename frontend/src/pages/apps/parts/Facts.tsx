import React from 'react';
import {Box, Stack, Typography} from '@mui/material';
import {CopyButton} from '../../../components/common';
import {FactSpec} from '../../../types/bundle.fact';
import {Chips} from './Chips';
import {ExpiryText} from './ExpiryText';
import {mono} from './mono';

const rowStyle = {
    display: 'grid', gridTemplateColumns: {xs: '1fr', sm: '180px 1fr'}, gap: {xs: 0, sm: 2}, py: 0.75,
    borderBottom: 1, borderColor: 'divider', '&:last-child': {borderBottom: 0},
};

const hasValue = (f: FactSpec): boolean => !!f.text || !!f.chips?.length || !!f.expiry;

const FactValue: React.FC<{ fact: FactSpec, now: number }> = ({fact, now}) => {
    if (fact.chips) return <Chips items={fact.chips} max={fact.max}/>;
    if (fact.expiry) return <ExpiryText date={fact.expiry} now={now}/>;
    return <Typography component="div" variant="body2" sx={fact.monospace ? mono : {wordBreak: 'break-word'}}>{fact.text}</Typography>;
};

const FactRow: React.FC<{ fact: FactSpec, now: number }> = ({fact, now}) =>
    <Box component="div" sx={rowStyle}>
        <Typography variant="body2" color="text.secondary">{fact.label}</Typography>
        <Stack direction="row" spacing={0.5} alignItems="center" sx={{minWidth: 0}}>
            <FactValue fact={fact} now={now}/>
            {fact.copy && fact.text && <CopyButton text={fact.text} label={`Copy ${fact.label}`}/>}
        </Stack>
    </Box>;

// Facts lists labelled values; the ones without a value are left out
export const Facts: React.FC<{ facts: FactSpec[], now: number }> = ({facts, now}) =>
    <>{facts.filter(hasValue).map(f => <FactRow key={f.label} fact={f} now={now}/>)}</>;
