import * as React from 'react';
import List from '@mui/material/List';
import InfoRoundedIcon from '@mui/icons-material/InfoRounded';
import KeyboardDoubleArrowLeftRoundedIcon from '@mui/icons-material/KeyboardDoubleArrowLeftRounded';
import KeyboardDoubleArrowRightRoundedIcon from '@mui/icons-material/KeyboardDoubleArrowRightRounded';
import MenuItemButton from './MenuItemButton';

interface MenuFooterProps {
    collapsed: boolean,
    // given only where the menu can collapse
    onToggleCollapsed?: () => void,
}

const ABOUT_URL = 'https://github.com/fsuhrau/automationhub';

// the bottom of the menu: the about link and the collapse switch
const MenuFooter: React.FC<MenuFooterProps> = ({collapsed, onToggleCollapsed}) => <List dense={true} sx={{p: 1}}>
    <MenuItemButton text="About" icon={<InfoRoundedIcon/>} collapsed={collapsed} href={ABOUT_URL}/>
    {onToggleCollapsed && <MenuItemButton text={collapsed ? 'Expand menu' : 'Collapse menu'} collapsed={collapsed}
        icon={collapsed ? <KeyboardDoubleArrowRightRoundedIcon/> : <KeyboardDoubleArrowLeftRoundedIcon/>}
        onClick={onToggleCollapsed}/>}
</List>;

export default MenuFooter;
