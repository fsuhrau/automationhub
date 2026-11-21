import React from 'react';
import {FormControl, IconButton, InputLabel, MenuItem, Select, Stack, Tooltip, Typography} from '@mui/material';
import {ArchiveOutlined, SwapHorizRounded} from '@mui/icons-material';
import {IAppBinaryData} from '../../../types/app';
import {ago, bundleLabel} from '../bundlesModel';

interface BundlePickerProps {
    which: 'a' | 'b',
    label: string,
    bundles: IAppBinaryData[],
    value?: number,
    now: number,
    onPick: (which: 'a' | 'b', id: number) => void,
}

const BundleOption: React.FC<{ bundle: IAppBinaryData, now: number }> = ({bundle, now}) =>
    <Stack direction="row" spacing={1} alignItems="center" sx={{minWidth: 0}}>
        <span>{bundleLabel(bundle)}</span>
        <Typography component="span" variant="caption" color="text.secondary">{ago(bundle.createdAt, now)}</Typography>
        {bundle.tags && <Typography component="span" variant="caption" color="text.secondary">{bundle.tags}</Typography>}
        {bundle.archivedAt && <ArchiveOutlined fontSize="inherit" color="disabled" aria-label="archived"/>}
    </Stack>;

const BundlePicker: React.FC<BundlePickerProps> = ({which, label, bundles, value, now, onPick}) =>
    <FormControl size="small" sx={{flex: 1, minWidth: 0}}>
        <InputLabel id={`compare-${which}`}>{label}</InputLabel>
        <Select labelId={`compare-${which}`} label={label} value={bundles.some(x => x.id === value) ? value : ''}
            onChange={e => onPick(which, Number(e.target.value))}>
            {bundles.map(x => <MenuItem key={x.id} value={x.id}><BundleOption bundle={x} now={now}/></MenuItem>)}
        </Select>
    </FormControl>;

interface ComparePickersProps {
    bundles: IAppBinaryData[],
    from?: number,
    to?: number,
    now: number,
    onPick: (which: 'a' | 'b', id: number) => void,
    onSwap: () => void,
}

// ComparePickers choose the two bundles to compare and swap them
export const ComparePickers: React.FC<ComparePickersProps> = ({bundles, from, to, now, onPick, onSwap}) =>
    <Stack direction={{xs: 'column', sm: 'row'}} spacing={1} alignItems={{sm: 'center'}}>
        <BundlePicker which="a" label="From" bundles={bundles} value={from} now={now} onPick={onPick}/>
        <Tooltip title="Swap">
            <IconButton aria-label="Swap the bundles" onClick={onSwap} sx={{alignSelf: 'center'}}><SwapHorizRounded/></IconButton>
        </Tooltip>
        <BundlePicker which="b" label="To" bundles={bundles} value={to} now={now} onPick={onPick}/>
    </Stack>;
