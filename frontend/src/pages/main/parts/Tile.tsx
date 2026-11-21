import React from 'react';
import {Box, ButtonBase, Card, Typography} from '@mui/material';
import {Link as RouterLink} from 'react-router-dom';

interface TileProps {
    label: string,
    value: React.ReactNode,
    note?: React.ReactNode,
    to?: string,
}

// a number with what it counts, leading to where it's managed
const Tile: React.FC<TileProps> = ({label, value, note, to}) => {
    const content = <Box sx={{p: 1.5, textAlign: 'left', width: '100%'}}>
        <Typography variant="caption" color="text.secondary">{label}</Typography>
        <Typography variant="h5" component="div">{value}</Typography>
        {note && <Typography variant="caption" color="text.secondary" component="div">{note}</Typography>}
    </Box>;
    return <Card variant="outlined" sx={{height: '100%'}}>
        {to ? <ButtonBase component={RouterLink} to={to} sx={{display: 'block', height: '100%', width: '100%'}}>{content}</ButtonBase> : content}
    </Card>;
};

export default Tile;
