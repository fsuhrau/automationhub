import IProject from '../../project/project';
import {IAppData} from '../../types/app';
import {NavItem} from './menuItems';

// what the open and the collapsed menu both show
export interface MenuViewProps {
    project: IProject,
    app: IAppData | null,
    projectItems: NavItem[],
    appItems: NavItem[],
    onNavigate?: () => void,
    onToggleCollapsed?: () => void,
    // opens the collapsed menu, e.g. to pick another app
    onExpand?: () => void,
}
