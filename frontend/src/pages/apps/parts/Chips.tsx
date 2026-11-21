import React, {useState} from 'react';
import {Chip, Stack} from '@mui/material';

interface ChipsProps {
    items?: string[],
    // chips shown before "+n more"
    max?: number,
}

// Chips lists short values; a long list opens on click
export const Chips: React.FC<ChipsProps> = ({items, max = 40}) => {
    const [all, setAll] = useState(false);
    if (!items || items.length === 0) return null;
    const shown = all ? items : items.slice(0, max);
    return <Stack direction="row" spacing={0.5} flexWrap="wrap" useFlexGap={true}>
        {shown.map(i => <Chip key={i} size="small" variant="outlined" label={i}/>)}
        {items.length > shown.length && <Chip size="small" label={`+${items.length - shown.length} more`} onClick={() => setAll(true)}/>}
    </Stack>;
};
