import React from 'react';
import {List, ListItemButton, ListItemIcon, ListItemText} from '@mui/material';
import PlatformTypeIcon from '../../../components/PlatformTypeIcon';
import {IAppData} from '../../../types/app';
import {getPlatformName} from '../../../types/platform.type.enum';

export interface AppListProps {
    apps: IAppData[],
    selectedId: number | undefined,
    onSelect: (id: number) => void,
}

const AppList: React.FC<AppListProps> = ({apps, selectedId, onSelect}) =>
    <List dense={true} sx={{border: 1, borderColor: 'divider', borderRadius: 1, p: 0.5, alignSelf: 'start'}} aria-label="Apps">
        {apps.map(a => <ListItemButton key={a.id} selected={a.id === selectedId} onClick={() => onSelect(a.id)} sx={{borderRadius: 1}}>
            <ListItemIcon sx={{minWidth: 36}}><PlatformTypeIcon platformType={a.platform} fontSize="small"/></ListItemIcon>
            <ListItemText primary={a.name} secondary={a.identifier || getPlatformName(a.platform)} slotProps={{secondary: {noWrap: true}}}/>
        </ListItemButton>)}
    </List>;

export default AppList;
