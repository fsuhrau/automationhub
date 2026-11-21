import React from 'react';
import {Box, Button, IconButton, TextField, Tooltip, Typography} from '@mui/material';
import {AddRounded, DeleteOutlineRounded} from '@mui/icons-material';
import {Section} from '../../../components/common';
import {ParameterRow} from '../deviceEditModel';

interface ParameterEditorProps {
    rows: ParameterRow[],
    errors: Map<number, string>,
    onChange: (rowId: string, change: Partial<ParameterRow>) => void,
    onAdd: () => void,
    onRemove: (rowId: string) => void,
}

interface RowFieldsProps extends Omit<ParameterEditorProps, 'rows' | 'onAdd'> {
    row: ParameterRow,
    index: number,
}

const noSpellcheck = {spellCheck: false};

const RowFields: React.FC<RowFieldsProps> = ({row, index, errors, onChange, onRemove}) => {
    const label = row.key || `parameter ${index + 1}`;
    return <>
        <TextField size="small" placeholder="Key" value={row.key} error={errors.has(index)} helperText={errors.get(index)}
                   // a row without a stored id was just added: put the cursor in it
            autoFocus={row.id === undefined}
            onChange={e => onChange(row.rowId, {key: e.target.value})}
            slotProps={{htmlInput: {...noSpellcheck, 'aria-label': `Key of parameter ${index + 1}`}}}/>
        <TextField size="small" placeholder="Value" value={row.value}
            onChange={e => onChange(row.rowId, {value: e.target.value})}
            slotProps={{htmlInput: {...noSpellcheck, 'aria-label': `Value of ${label}`}}}/>
        <Tooltip title="Remove">
            <IconButton aria-label={`Remove ${label}`} onClick={() => onRemove(row.rowId)}>
                <DeleteOutlineRounded fontSize="small"/>
            </IconButton>
        </Tooltip>
    </>;
};

const grid = {
    display: 'grid', gridTemplateColumns: {xs: '1fr', sm: 'minmax(140px, 2fr) 3fr auto'}, gap: 1, alignItems: 'start', mb: 1.5,
};

const Rows: React.FC<ParameterEditorProps> = ({rows, ...rest}) =>
    <Box sx={grid} role="group" aria-label="Custom parameters">
        {rows.map((row, i) => <RowFields key={row.rowId} row={row} index={i} {...rest}/>)}
    </Box>;

const description = <>
    Values the hub uses for this device and passes to tests, for example <Box component="span" sx={{fontFamily: 'monospace'}}>PIN</Box> to
    unlock an Android phone. Keys must be unique.
</>;

// ParameterEditor edits the custom parameters of a device, row by row
export const ParameterEditor: React.FC<ParameterEditorProps> = props =>
    <Section title="Custom parameters" description={description}>
        {props.rows.length === 0
            ? <Typography variant="body2" color="text.secondary" sx={{mb: 1.5}}>No custom parameters.</Typography>
            : <Rows {...props}/>}
        <Button size="small" startIcon={<AddRounded/>} onClick={props.onAdd}>Add parameter</Button>
    </Section>;
