import {useMemo, useState} from 'react';
import {DeviceControlApi, Point} from '../../../types/device.control.api';
import {SceneGraph} from '../../../types/device.control.scene';
import {errorMessage} from '../../../lib/errorMessage';
import {elementAt, highlightStyle, parseSceneGraph, scenePointAt} from './sceneModel';
import {Runner} from './useGuardedAction';
import {SceneSelection, useSceneSelection} from './useSceneSelection';

export interface SceneControl extends SceneSelection {
    scene: SceneGraph | null,
    // why the scene graph couldn't be loaded; shown in the tab
    error: string | null,
    // clicks on the screen select an element instead of touching the device
    picking: boolean,
    togglePicking: () => void,
    load: () => void,
    // selects the element under a point of the screen element
    pickAt: (el: HTMLElement, client: Point) => void,
    // where the selected element is on the screen, in percent
    highlight: Record<string, string> | null,
}

// useSceneGraph loads the scene graph of the app running on the device and
// keeps the selection and the pick mode. A missing app is an expected
// answer here, so a failure shows in the tab instead of as an error.
export const useSceneGraph = (service: DeviceControlApi, run: Runner): SceneControl => {
    const [scene, setScene] = useState<SceneGraph | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [picking, setPicking] = useState(false);
    const selection = useSceneSelection(scene);
    const {selected, select, reset} = selection;

    const load = (): void => {
        setError(null);
        run(async () => {
            const graph = parseSceneGraph(await service.hierarchy());
            setScene(graph);
            reset(graph);
        }, e => setError(errorMessage(e)));
    };
    const pickAt = (el: HTMLElement, client: Point): void => {
        if (!scene || !scene.width || !scene.height) return;
        select(elementAt(scene.root, scenePointAt(el.getBoundingClientRect(), scene, client)));
    };
    const highlight = useMemo(() => highlightStyle(selected, scene), [selected, scene]);
    return {...selection, scene, error, picking, togglePicking: () => setPicking(p => !p), load, pickAt, highlight};
};
