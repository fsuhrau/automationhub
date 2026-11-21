import React from 'react';
import {Box, IconButton, Table, TableBody, TableCell, TableHead, TableRow, Tooltip, Typography} from '@mui/material';
import {DeleteOutlineRounded, EditRounded} from '@mui/icons-material';
import {AppParameter} from '../../../types/app';
import {parameterChoices} from '../settingsModel';

export interface ParametersTableProps {
    parameters: AppParameter[],
    onEdit: (parameter: AppParameter) => void,
    onDelete: (parameter: AppParameter) => void,
}

const KindCell: React.FC<{ parameter: AppParameter }> = ({parameter}) => {
    const choices = parameterChoices(parameter);
    return <TableCell>
        {parameter.type.type === 'option'
            ? <Tooltip title={choices.join(', ')}><span>Choice of {choices.length}</span></Tooltip>
            : 'Text'}
    </TableCell>;
};

const ParameterRow: React.FC<{ parameter: AppParameter } & Omit<ParametersTableProps, 'parameters'>> = ({parameter, onEdit, onDelete}) =>
    <TableRow>
        <TableCell sx={{fontFamily: 'monospace'}}>{parameter.name}</TableCell>
        <KindCell parameter={parameter}/>
        <TableCell sx={{wordBreak: 'break-all'}}>
            {parameter.type.defaultValue || <Typography variant="body2" color="text.secondary" component="span">none</Typography>}
        </TableCell>
        <TableCell align="right" sx={{whiteSpace: 'nowrap'}}>
            <Tooltip title="Edit">
                <IconButton size="small" aria-label={`Edit ${parameter.name}`} onClick={() => onEdit(parameter)}><EditRounded fontSize="small"/></IconButton>
            </Tooltip>
            <Tooltip title="Delete">
                <IconButton size="small" aria-label={`Delete ${parameter.name}`} onClick={() => onDelete(parameter)}>
                    <DeleteOutlineRounded fontSize="small"/>
                </IconButton>
            </Tooltip>
        </TableCell>
    </TableRow>;

const ParametersTable: React.FC<ParametersTableProps> = ({parameters, ...actions}) =>
    <Box sx={{overflowX: 'auto'}}><Table size="small" aria-label="Parameters">
        <TableHead><TableRow>
            <TableCell>Name</TableCell><TableCell>Kind</TableCell><TableCell>Default</TableCell><TableCell/>
        </TableRow></TableHead>
        <TableBody>{parameters.map(p => <ParameterRow key={p.id} parameter={p} {...actions}/>)}</TableBody>
    </Table></Box>;

export default ParametersTable;
