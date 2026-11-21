import React from 'react';
import {Chip, Stack} from '@mui/material';

interface DeviceChipsProps {
    ids: number[],
    selected: number | null,
    nameOf: (id: number) => string,
    onPick: (id: number) => void,
}

export const DeviceChips: React.FC<DeviceChipsProps> = ({ids, selected, nameOf, onPick}) =>
    <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap={true} sx={{mb: 1.5}} role="group" aria-label="Device">
        {ids.map(id => <Chip key={id} size="small" label={nameOf(id)}
            color={id === selected ? 'primary' : 'default'}
            variant={id === selected ? 'filled' : 'outlined'}
            onClick={() => onPick(id)}/>)}
    </Stack>;
