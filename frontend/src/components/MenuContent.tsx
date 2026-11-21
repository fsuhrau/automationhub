import * as React from 'react';
import {useLocation} from "react-router-dom";
import {useProjectContext} from "../hooks/ProjectProvider";
import {useSelectedApp} from "../hooks/useSelectedApp";
import CollapsedMenu from "./menu/CollapsedMenu";
import ExpandedMenu from "./menu/ExpandedMenu";
import {appNavItems, projectNavItems} from "./menu/menuItems";

interface MenuContentProps {
    onNavigate?: () => void,
    collapsed?: boolean,
    onToggleCollapsed?: () => void,
    // opens the collapsed menu, e.g. to pick another app
    onExpand?: () => void,
}

// MenuContent is the project navigation. It is split by scope, so it's clear
// what a page belongs to: the project pages (overview, devices, settings) are
// the same for every app, while tests and bundles belong to the app picked in
// the "App" section. Collapsed (onToggleCollapsed is given only where the
// menu can collapse) it is a rail of icons.
export default function MenuContent({collapsed = false, ...handlers}: MenuContentProps): React.ReactElement {
    const {project, projectIdentifier} = useProjectContext();
    const {app} = useSelectedApp();
    const {pathname} = useLocation();

    const view = {
        project, app, ...handlers,
        projectItems: projectNavItems(projectIdentifier, pathname),
        appItems: appNavItems(projectIdentifier, app, pathname),
    };
    return collapsed ? <CollapsedMenu {...view}/> : <ExpandedMenu {...view}/>;
}
