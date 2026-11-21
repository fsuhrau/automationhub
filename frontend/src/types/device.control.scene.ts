import {TreeViewBaseItem} from '@mui/x-tree-view/models';

// The scene graph of the app running on a device, parsed from its XML.
export interface SceneNode extends TreeViewBaseItem {
    element: Element,
}

export interface SceneGraph {
    items: SceneNode[],
    root: Element,
    byId: Map<string, Element>,
    idOf: Map<Element, string>,
    // the size of the scene in scene units
    width: number,
    height: number,
}
