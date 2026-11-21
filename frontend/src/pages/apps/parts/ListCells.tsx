import React, {useState} from 'react';
import {Chip, Stack, TableCell, Typography} from '@mui/material';
import {ListRow} from '../../../types/bundle.compare';

const MAX = 30;

interface EntriesProps {
    items: string[],
    sign: string,
    color: 'success' | 'error',
    all: boolean,
    onAll: () => void,
}

const ellipsis = {maxWidth: '100%', '& .MuiChip-label': {overflow: 'hidden', textOverflow: 'ellipsis'}};

const Entries: React.FC<EntriesProps> = ({items, sign, color, all, onAll}) => {
    const shown = all ? items : items.slice(0, MAX);
    return <>
        {shown.map(i => <Chip key={sign + i} size="small" color={color} variant="outlined" label={`${sign} ${i}`} sx={ellipsis}/>)}
        {items.length > shown.length && <Chip size="small" label={`+${items.length - shown.length} more`} onClick={onAll}/>}
    </>;
};

const unchanged = (row: ListRow): string => {
    if (row.change === 'same') return `${row.same} the same in both`;
    return row.same > 0 ? `${row.same} unchanged` : '';
};

// ListCells shows what only one of the bundles has; a long list opens on click
export const ListCells: React.FC<{ row: ListRow }> = ({row}) => {
    const [all, setAll] = useState(false);
    const onAll = (): void => setAll(true);
    return <TableCell colSpan={2} sx={{verticalAlign: 'top'}}>
        <Stack direction="row" spacing={0.5} flexWrap="wrap" useFlexGap={true} alignItems="center">
            <Entries items={row.added} sign="+" color="success" all={all} onAll={onAll}/>
            <Entries items={row.removed} sign="−" color="error" all={all} onAll={onAll}/>
            <Typography variant="caption" color="text.secondary" sx={{ml: 0.5}}>{unchanged(row)}</Typography>
        </Stack>
    </TableCell>;
};
