import React from 'react';
import {Box, Stack, TableCell, Typography} from '@mui/material';
import {ExpandLessRounded, ExpandMoreRounded} from '@mui/icons-material';
import {deviceLabel} from '../runStatusModel';
import {TestRow} from '../../testRunModel';

// breakAfterDots lets long dotted names wrap after a dot, not mid-word
const breakAfterDots = (name: string): React.ReactNode =>
    name.split('.').map((part, i, all) => <React.Fragment key={all.slice(0, i + 1).join('.')}>{part}{i < all.length - 1 && <>.<wbr/></>}</React.Fragment>);

const Caption: React.FC<{ children: React.ReactNode }> = ({children}) =>
    <Typography variant="caption" color="text.secondary" component="div">{children}</Typography>;

const DeviceCaption: React.FC<{ row: TestRow }> = ({row}) => {
    const device = row.protocol.device;
    if (!device) return null;
    return <Caption>{deviceLabel(device)} · {device.os} {device.osVersion}</Caption>;
};

interface TestNameCellProps {
    row: TestRow,
    hasChildren: boolean,
    expanded: boolean,
}

export const TestNameCell: React.FC<TestNameCellProps> = ({row, hasChildren, expanded}) =>
    <TableCell sx={{minWidth: 0, px: 1}}>
        <Stack direction="row" alignItems="center" spacing={0.5}>
            {hasChildren && (expanded ? <ExpandLessRounded fontSize="small"/> : <ExpandMoreRounded fontSize="small"/>)}
            <Box sx={{minWidth: 0}}>
                <Typography variant="body2" sx={{fontWeight: 600, overflowWrap: 'break-word'}}>
                    {row.name}
                    {hasChildren && <Typography component="span" variant="caption" color="text.secondary" sx={{ml: 1}}>
                        {row.children.length} tests
                    </Typography>}
                </Typography>
                {row.suite && <Caption>{breakAfterDots(row.suite)}</Caption>}
                <DeviceCaption row={row}/>
            </Box>
        </Stack>
    </TableCell>;
