import React from 'react';
import {FormControl, InputLabel, MenuItem, Select, Stack, TextField, Typography} from '@mui/material';
import PlatformTypeIcon from '../../../components/PlatformTypeIcon';
import {getPlatformTypes, PlatformType} from '../../../types/platform.type.enum';
import {asksIdentifier, identifierLabel, platformLabel} from '../settingsModel';

export const PlatformSelect: React.FC<{ value: PlatformType, onChange: (p: PlatformType) => void }> = ({value, onChange}) =>
    <FormControl size="small">
        <InputLabel id="new-app-platform">Platform</InputLabel>
        <Select labelId="new-app-platform" label="Platform" value={value} onChange={e => onChange(Number(e.target.value) as PlatformType)}>
            {getPlatformTypes().map(p => <MenuItem key={p.id} value={Number(p.id)}>
                <Stack direction="row" spacing={1} alignItems="center">
                    <PlatformTypeIcon platformType={Number(p.id) as PlatformType} fontSize="small"/>
                    <span>{platformLabel(Number(p.id) as PlatformType)}</span>
                </Stack>
            </MenuItem>)}
        </Select>
    </FormControl>;

export interface IdentifierFieldProps {
    platform: PlatformType,
    value: string,
    error?: string,
    onChange: (value: string) => void,
}

// IdentifierField asks for the identifier where it is used, and says
// otherwise that the platform needs none
export const IdentifierField: React.FC<IdentifierFieldProps> = ({platform, value, error, onChange}) => {
    if (!asksIdentifier(platform)) {
        return <Typography variant="body2" color="text.secondary">
            {platform === PlatformType.Editor ? 'Unity editor apps run in the editor and need no bundle.' : 'Web apps run in a browser and need no bundle.'}
        </Typography>;
    }
    return <TextField size="small" label={identifierLabel(platform)} placeholder="com.example.game" value={value}
        onChange={e => onChange(e.target.value)} error={!!error} helperText={error ?? 'The hub starts and stops the app by it.'}/>;
};
