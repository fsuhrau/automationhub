// The scene graph of the app on the device: parsing its XML, the XPath of an
// element, picking an element by position and where it is on the screen.
import {SceneGraph, SceneNode} from '../../../types/device.control.scene';
import {Point} from '../../../types/device.control.api';

export interface SceneElement {
    getAttribute(name: string): string | null,

    readonly children: ArrayLike<SceneElement>,
}

export interface Rect {
    left: number,
    top: number,
    width: number,
    height: number,
}

const attr = (e: SceneElement, name: string): number => Number(e.getAttribute(name)) || 0;

export const nodeLabel = (e: Element): string => {
    const name = e.getAttribute('Name');
    const text = e.getAttribute('LabelText');
    return `${e.nodeName}${name ? ` "${name}"` : ''}${text ? ` – ${text}` : ''}`;
};

export const parseSceneGraph = (xml: string): SceneGraph => {
    const root = new DOMParser().parseFromString(xml, 'text/xml').documentElement;
    if (!root || root.nodeName === 'parsererror') {
        throw new Error('scene graph is not valid XML');
    }
    const byId = new Map<string, Element>();
    const idOf = new Map<Element, string>();
    const build = (e: Element): SceneNode => {
        const id = `n${byId.size}`;
        byId.set(id, e);
        idOf.set(e, id);
        return {id, label: nodeLabel(e), element: e, children: Array.from(e.children).map(build)};
    };
    return {items: [build(root)], root, byId, idOf, width: attr(root, 'RectangleX'), height: attr(root, 'RectangleY')};
};

const positionIn = (e: Element): number =>
    Array.from(e.parentElement?.children ?? []).filter(c => c.nodeName === e.nodeName).indexOf(e) + 1;

// xpathFor names an element by its Name attribute, else by its position
// below the root
export const xpathFor = (e: Element): string => {
    const name = e.getAttribute('Name');
    if (name) {
        return `//${e.nodeName}[@Name='${name.replace(/'/g, "\\'")}']`;
    }
    const parts: string[] = [];
    let current: Element = e;
    while (current.parentElement) {
        parts.unshift(`${current.nodeName}[${positionIn(current)}]`);
        current = current.parentElement;
    }
    return `/${current.nodeName}${parts.length ? `/${parts.join('/')}` : ''}`;
};

// the ids of an element's ancestors, to expand the tree down to it
export const ancestorIds = (scene: SceneGraph, e: Element): string[] => {
    const ids: string[] = [];
    for (let current = e.parentElement; current; current = current.parentElement) {
        const id = scene.idOf.get(current);
        if (id) ids.unshift(id);
    }
    return ids;
};

const inside = (e: SceneElement, p: Point): boolean => {
    const [x, y, w, h] = [attr(e, 'X'), attr(e, 'Y'), attr(e, 'RectangleX'), attr(e, 'RectangleY')];
    return w > 0 && h > 0 && p.x >= x && p.y >= y && p.x < x + w && p.y < y + h;
};

// elementAt finds the deepest visible element at a point in scene
// coordinates; later siblings are drawn on top, so they win
export const elementAt = <E extends SceneElement>(root: E, p: Point): E | null => {
    const kids = Array.from(root.children) as E[];
    for (let i = kids.length - 1; i >= 0; i--) {
        const hit = elementAt(kids[i], p);
        if (hit) return hit;
    }
    return inside(root, p) ? root : null;
};

// scenePointAt turns a pointer position over the screen element into scene
// coordinates
export const scenePointAt = (rect: Rect, scene: Pick<SceneGraph, 'width' | 'height'>, client: Point): Point => ({
    x: ((client.x - rect.left) / rect.width) * scene.width,
    y: ((client.y - rect.top) / rect.height) * scene.height,
});

// highlightStyle places the selected element over the screen, in percent
export const highlightStyle = (selected: SceneElement | null, scene: Pick<SceneGraph, 'width' | 'height'> | null): Record<string, string> | null => {
    if (!selected || !scene || !scene.width || !scene.height) return null;
    const pct = (name: string, of: number): string => `${(attr(selected, name) / of) * 100}%`;
    return {left: pct('X', scene.width), top: pct('Y', scene.height), width: pct('RectangleX', scene.width), height: pct('RectangleY', scene.height)};
};
