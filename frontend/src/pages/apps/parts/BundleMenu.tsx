import React from 'react';
import {ListItemIcon, ListItemText, Menu, MenuItem} from '@mui/material';
import {ArchiveOutlined, CompareArrowsRounded, ContentCopyRounded, DeleteOutlineRounded, InfoOutlined} from '@mui/icons-material';
import {Link as RouterLink} from 'react-router-dom';
import {IAppBinaryData} from '../../../types/app';

export interface MenuTarget {
    anchor: HTMLElement,
    bundle: IAppBinaryData,
}

interface BundleMenuProps {
    target: MenuTarget,
    detailUrl: string,
    // the compare page with this bundle picked
    compareUrl: string,
    // a bundle older than this one exists
    hasPrevious: boolean,
    onCopyHash: () => void,
    onArchive: () => void,
    onDelete: () => void,
    onClose: () => void,
}

const Item: React.FC<{ icon: React.ReactNode, primary: string, secondary?: string }> = ({icon, primary, secondary}) =>
    <><ListItemIcon>{icon}</ListItemIcon><ListItemText primary={primary} secondary={secondary}/></>;

// BundleMenu is the menu of a row: details, compare, copy the hash, archive, delete
export const BundleMenu: React.FC<BundleMenuProps> = props => {
    const {target, detailUrl, compareUrl, hasPrevious, onCopyHash, onArchive, onDelete, onClose} = props;
    const {bundle} = target;
    return <Menu anchorEl={target.anchor} open={true} onClose={onClose}>
        <MenuItem component={RouterLink} to={detailUrl}><Item icon={<InfoOutlined fontSize="small"/>} primary="Details"/></MenuItem>
        <MenuItem component={RouterLink} to={compareUrl} disabled={!hasPrevious}>
            <Item icon={<CompareArrowsRounded fontSize="small"/>} primary="Compare with previous"/>
        </MenuItem>
        <MenuItem disabled={!bundle.hash} onClick={onCopyHash}>
            <Item icon={<ContentCopyRounded fontSize="small"/>} primary="Copy hash" secondary={bundle.hash ? `${bundle.hash.slice(0, 12)}…` : undefined}/>
        </MenuItem>
        <MenuItem disabled={!!bundle.archivedAt} onClick={onArchive}>
            <Item icon={<ArchiveOutlined fontSize="small"/>} primary={bundle.archivedAt ? 'Archived' : 'Archive bundle'}
                secondary={bundle.archivedAt ? undefined : 'Remove the file, keep the details'}/>
        </MenuItem>
        <MenuItem onClick={onDelete} sx={{color: 'error.main'}}>
            <Item icon={<DeleteOutlineRounded fontSize="small" color="error"/>} primary="Delete bundle"/>
        </MenuItem>
    </Menu>;
};
