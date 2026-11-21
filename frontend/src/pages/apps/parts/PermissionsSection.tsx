import React from 'react';
import {Chip, Stack, Typography} from '@mui/material';
import {Section} from '../../../components/common';
import {AndroidFeature, AndroidInfo, AndroidPermission} from '../../../types/bundle.info';
import {isDangerous, shortPermission} from '../bundleDetailModel';
import {mono} from './mono';

const PermissionList: React.FC<{ permissions: AndroidPermission[] }> = ({permissions}) =>
    permissions.length === 0
        ? <Typography variant="body2" color="text.secondary">None requested.</Typography>
        : <Stack spacing={0.25} sx={{mb: 2}}>
            {permissions.map(p => <Stack key={p.name} direction="row" spacing={1} alignItems="center">
                <Typography variant="body2" sx={mono}>{shortPermission(p.name)}</Typography>
                {isDangerous(p.name) && <Chip size="small" variant="outlined" color="warning" label="runtime"/>}
                {p.maxSdk && <Typography variant="caption" color="text.secondary">up to API {p.maxSdk}</Typography>}
            </Stack>)}
        </Stack>;

const FeatureList: React.FC<{ features: AndroidFeature[] }> = ({features}) => <>
    <Typography variant="subtitle2" sx={{mt: 1, mb: 0.5}}>Hardware features ({features.length})</Typography>
    <Stack spacing={0.25}>
        {features.map(f => <Stack key={f.name} direction="row" spacing={1} alignItems="center">
            <Typography variant="body2" sx={mono}>{f.name.replace(/^android\.hardware\./, '')}</Typography>
            <Typography variant="caption" color="text.secondary">{f.required ? 'required' : 'optional'}</Typography>
        </Stack>)}
    </Stack>
</>;

export const PermissionsSection: React.FC<{ android: AndroidInfo }> = ({android}) => {
    const permissions = android.permissions ?? [];
    const features = android.features ?? [];
    const dangerous = permissions.filter(p => isDangerous(p.name)).length;
    return <Section title={`Permissions (${permissions.length})`} fullHeight={true}
        description={dangerous > 0 ? `${dangerous} need the player's consent at runtime` : undefined}>
        <PermissionList permissions={permissions}/>
        {features.length > 0 && <FeatureList features={features}/>}
    </Section>;
};
