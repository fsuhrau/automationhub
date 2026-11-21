import * as React from 'react';
import DashboardRoundedIcon from '@mui/icons-material/DashboardRounded';
import AnalyticsRoundedIcon from '@mui/icons-material/AnalyticsRounded';
import Inventory2RoundedIcon from '@mui/icons-material/Inventory2Rounded';
import SettingsRoundedIcon from '@mui/icons-material/SettingsRounded';
import DevicesRoundedIcon from '@mui/icons-material/DevicesRounded';
import {routes} from '../../lib/routes';
import {IAppData} from '../../types/app';

export interface NavItem {
    text: string,
    ref: string,
    icon: React.ReactElement,
    selected: boolean,
    hint?: string,
}

// the pages of the project, the same for every app
export const projectNavItems = (identifier: string, pathname: string): NavItem[] => {
    const project = routes.project(identifier);
    const overview = project.overview();
    return [
        {text: 'Overview', ref: overview, icon: <DashboardRoundedIcon/>, selected: pathname === overview || pathname.endsWith('/home')},
        {text: 'Device Pool', ref: project.devices(), icon: <DevicesRoundedIcon/>, selected: pathname.startsWith(`${overview}/device`)},
        {text: 'Settings', ref: project.settings(), icon: <SettingsRoundedIcon/>, selected: pathname.startsWith(project.settings())},
    ];
};

// the pages of the app that is picked; none without an app
export const appNavItems = (identifier: string, app: IAppData | null, pathname: string): NavItem[] => {
    if (!app) return [];
    const pages = routes.project(identifier).app(app.id);
    // ".../test" also covers ".../tests" and the pages of one test
    const testsPrefix = pages.tests().slice(0, -1);
    return [
        {text: 'Tests', ref: pages.tests(), icon: <AnalyticsRoundedIcon/>, selected: pathname.startsWith(testsPrefix)},
        {text: 'Bundles', ref: pages.bundles(), icon: <Inventory2RoundedIcon/>, selected: pathname.startsWith(pages.bundles())},
    ];
};
