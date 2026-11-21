import {useState} from 'react';
import {SceneGraph} from '../../../types/device.control.scene';
import {ancestorIds} from './sceneModel';

export interface SceneSelection {
    selected: Element | null,
    expanded: string[],
    setExpanded: (ids: string[]) => void,
    // selects an element and opens the tree down to it
    select: (element: Element | null) => void,
    // a new scene graph: nothing selected, the root open
    reset: (scene: SceneGraph) => void,
}

// useSceneSelection keeps the selected element of the scene graph and which
// tree nodes are open.
export const useSceneSelection = (scene: SceneGraph | null): SceneSelection => {
    const [selected, setSelected] = useState<Element | null>(null);
    const [expanded, setExpanded] = useState<string[]>([]);
    const select = (element: Element | null): void => {
        setSelected(element);
        if (element && scene) setExpanded(prev => Array.from(new Set([...prev, ...ancestorIds(scene, element)])));
    };
    const reset = (next: SceneGraph): void => {
        setSelected(null);
        setExpanded(next.items.map(i => i.id));
    };
    return {selected, expanded, setExpanded, select, reset};
};
