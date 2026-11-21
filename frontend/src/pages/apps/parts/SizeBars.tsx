import React from 'react';
import {Box, Tooltip, Typography} from '@mui/material';
import {Contents, SizeEntry} from '../../../types/bundle.info';
import {byteFormat} from '../../tests/value_formatter';
import {barWidth, shareOf} from '../bundleDetailModel';
import {mono} from './mono';

const rowStyle = {
    display: 'grid', gridTemplateColumns: {xs: '1fr 80px', sm: 'minmax(120px, 280px) 1fr 80px'}, gap: 1.5,
    alignItems: 'center', py: 0.5, '&:hover': {bgcolor: 'action.hover'}, borderRadius: 1, px: 0.5,
};

const Details: React.FC<{ part: SizeEntry, total: number }> = ({part, total}) => <>
    <div>{part.path}</div>
    <div>{byteFormat(part.size)} installed · {byteFormat(part.compressed)} in the file</div>
    <div>{part.files} {part.files === 1 ? 'file' : 'files'} · {(shareOf(part.size, total) * 100).toFixed(1)} %</div>
</>;

interface SizeBarProps {
    part: SizeEntry,
    max: number,
    total: number,
}

const SizeBar: React.FC<SizeBarProps> = ({part, max, total}) =>
    <Tooltip placement="top-start" title={<Details part={part} total={total}/>}>
        <Box component="li" sx={rowStyle}>
            <Typography variant="body2" sx={{...mono, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', wordBreak: 'normal'}}>
                {part.path}
            </Typography>
            <Box sx={{display: {xs: 'none', sm: 'block'}, height: 10}}>
                <Box sx={{height: '100%', width: `${barWidth(part.size, max)}%`, bgcolor: 'primary.main', borderRadius: '0 4px 4px 0'}}/>
            </Box>
            <Typography variant="body2" color="text.secondary" align="right" sx={{whiteSpace: 'nowrap'}}>{byteFormat(part.size)}</Typography>
        </Box>
    </Tooltip>;

// SizeBars shows where the size of a bundle goes: one bar per part, longest first
export const SizeBars: React.FC<{ contents: Contents }> = ({contents}) => {
    const groups = contents.groups ?? [];
    const max = Math.max(1, ...groups.map(g => g.size));
    return <Box component="ul" aria-label="Size by part" sx={{listStyle: 'none', p: 0, m: 0}}>
        {groups.map(g => <SizeBar key={g.path} part={g} max={max} total={contents.size}/>)}
    </Box>;
};
