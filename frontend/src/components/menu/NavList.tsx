import * as React from 'react';
import {useNavigate} from 'react-router-dom';
import MenuItemButton from './MenuItemButton';
import {NavItem} from './menuItems';

interface NavListProps {
    items: NavItem[],
    collapsed?: boolean,
    onNavigate?: () => void,
}

const NavList: React.FC<NavListProps> = ({items, collapsed, onNavigate}) => {
    const navigate = useNavigate();
    const open = (ref: string) => (): void => {
        navigate(ref);
        onNavigate?.();
    };
    return <>
        {items.map(item => (
            <MenuItemButton key={item.text} text={item.text} hint={item.hint} icon={item.icon}
                collapsed={collapsed} selected={item.selected} onClick={open(item.ref)}/>
        ))}
    </>;
};

export default NavList;
