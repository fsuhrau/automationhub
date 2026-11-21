import {useCallback, useState} from 'react';

const STORAGE_KEY = 'mainMenuCollapsed';

const loadCollapsed = (): boolean => {
    try {
        return localStorage.getItem(STORAGE_KEY) === 'true';
    } catch {
        return false;
    }
};

const saveCollapsed = (collapsed: boolean): void => {
    try {
        localStorage.setItem(STORAGE_KEY, `${collapsed}`);
    } catch {
        // the menu still collapses, it just isn't remembered
    }
};

export interface MenuCollapse {
    collapsed: boolean,
    toggle: () => void,
    expand: () => void,
}

// whether the menu is a rail of icons; the choice is remembered in the browser
export const useMenuCollapsed = (): MenuCollapse => {
    const [collapsed, setCollapsed] = useState(loadCollapsed);
    const set = useCallback((value: boolean) => {
        saveCollapsed(value);
        setCollapsed(value);
    }, []);
    const toggle = useCallback(() => set(!collapsed), [set, collapsed]);
    const expand = useCallback(() => set(false), [set]);
    return {collapsed, toggle, expand};
};
