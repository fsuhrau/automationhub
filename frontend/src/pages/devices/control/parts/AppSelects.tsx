import React from 'react';
import {Divider, FormControl, InputLabel, ListItemText, MenuItem, Select} from '@mui/material';
import {formatDateTime} from '../../../../components/DateTime';
import {AppPicker} from '../useAppPicker';

const toId = (value: unknown): number | '' => String(value) === '' ? '' : Number(value);

export const AppSelect: React.FC<{ picker: AppPicker }> = ({picker}) => (
    <FormControl size="small" fullWidth={true}>
        <InputLabel id="control-app">App</InputLabel>
        <Select labelId="control-app" label="App" value={picker.appId} onChange={e => picker.chooseApp(toId(e.target.value))}>
            {picker.fittingApps.map(a => <MenuItem key={a.id} value={a.id}>{a.name}</MenuItem>)}
            {picker.otherApps.length > 0 && <Divider/>}
            {picker.otherApps.map(a => <MenuItem key={a.id} value={a.id}>
                <ListItemText primary={a.name} secondary="other platform"/>
            </MenuItem>)}
        </Select>
    </FormControl>
);

export const BundleSelect: React.FC<{ picker: AppPicker }> = ({picker}) => {
    const {bundles} = picker;
    const label = (id: unknown): string => {
        const b = bundles.find(x => x.id === id);
        return b ? `${b.version || b.name} (#${b.id})` : '';
    };
    return (
        <FormControl size="small" fullWidth={true} disabled={picker.appId === '' || bundles.length === 0}>
            <InputLabel id="control-bundle">Bundle</InputLabel>
            <Select labelId="control-bundle" label="Bundle" value={picker.bundleId} renderValue={label}
                onChange={e => picker.chooseBundle(toId(e.target.value))}>
                {bundles.map((b, i) => <MenuItem key={b.id} value={b.id}>
                    <ListItemText primary={`${b.version || b.name}${i === 0 ? ' (newest)' : ''}`}
                        secondary={`#${b.id} · uploaded ${formatDateTime(b.createdAt, 'YYYY/MM/DD HH:mm')}`}/>
                </MenuItem>)}
            </Select>
        </FormControl>
    );
};
