import React from 'react';
import {Box, Chip, IconButton, Stack, Table, TableBody, TableCell, TableHead, TableRow, Tooltip, Typography} from '@mui/material';
import {DeleteOutlineRounded, KeyRounded} from '@mui/icons-material';
import {CopyButton} from '../../../components/common';
import {INodeData} from '../../../types/node';

export interface NodesTableProps {
    nodes: INodeData[],
    onReissue: (node: INodeData) => void,
    onDelete: (node: INodeData) => void,
}

const CONNECTED = 1;

const NodeRow: React.FC<{ node: INodeData } & Omit<NodesTableProps, 'nodes'>> = ({node, onReissue, onDelete}) =>
    <TableRow>
        <TableCell>{node.name}</TableCell>
        <TableCell>
            {node.status === CONNECTED
                ? <Chip size="small" color="success" label="Connected"/>
                : <Chip size="small" label="Offline"/>}
        </TableCell>
        <TableCell sx={{whiteSpace: 'nowrap'}}>
            <Stack direction="row" alignItems="center" spacing={0.5}>
                <Typography variant="body2" sx={{fontFamily: 'monospace'}}>{node.identifier}</Typography>
                <CopyButton text={node.identifier} label="Copy the identifier"/>
            </Stack>
        </TableCell>
        <TableCell align="right" sx={{whiteSpace: 'nowrap'}}>
            <Tooltip title="New token, to set the node up again">
                <IconButton size="small" aria-label={`New token for ${node.name}`} onClick={() => onReissue(node)}><KeyRounded fontSize="small"/></IconButton>
            </Tooltip>
            <Tooltip title="Delete">
                <IconButton size="small" aria-label={`Delete ${node.name}`} onClick={() => onDelete(node)}><DeleteOutlineRounded fontSize="small"/></IconButton>
            </Tooltip>
        </TableCell>
    </TableRow>;

const NodesTable: React.FC<NodesTableProps> = ({nodes, ...actions}) =>
    <Box sx={{overflowX: 'auto'}}><Table size="small" aria-label="Nodes">
        <TableHead><TableRow>
            <TableCell>Name</TableCell><TableCell>Status</TableCell><TableCell>Identifier</TableCell><TableCell/>
        </TableRow></TableHead>
        <TableBody>{nodes.map(n => <NodeRow key={n.id} node={n} {...actions}/>)}</TableBody>
    </Table></Box>;

export default NodesTable;
