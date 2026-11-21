import {describe, expect, it} from 'vitest';
import {SceneGraph} from '../../../types/device.control.scene';
import {
    ancestorIds, elementAt, highlightStyle, nodeLabel, parseSceneGraph, scenePointAt, xpathFor,
} from './sceneModel';

const XML = `<Scene RectangleX="200" RectangleY="400" Name="Root">
  <Panel X="0" Y="0" RectangleX="200" RectangleY="200">
    <Button ID="b1" Name="Play" LabelText="Go" X="10" Y="10" RectangleX="50" RectangleY="20"/>
    <Button ID="b2" Name="O'Neil" X="100" Y="10" RectangleX="50" RectangleY="20"/>
  </Panel>
  <Panel X="0" Y="200" RectangleX="200" RectangleY="200"><Label/></Panel>
</Scene>`;

const parsed = (): SceneGraph => parseSceneGraph(XML);
const find = (name: string): Element => {
    const root = parsed().root;
    return root.getAttribute('Name') === name ? root : root.querySelector(`[Name="${name}"]`) as Element;
};

describe('parseSceneGraph', () => {
    it('builds the tree with ids and the scene size', () => {
        const scene = parsed();
        expect(scene.width).toBe(200);
        expect(scene.height).toBe(400);
        expect(scene.items).toHaveLength(1);
        expect(scene.items[0].children?.map(c => c.id)).toEqual(['n1', 'n4']);
        expect(scene.byId.get('n2')?.getAttribute('ID')).toBe('b1');
        expect(scene.idOf.get(scene.root)).toBe('n0');
    });

    it('takes a missing size as zero', () => {
        expect(parseSceneGraph('<Scene/>').width).toBe(0);
    });

    it('rejects text that is not XML', () => {
        expect(() => parseSceneGraph('<<<')).toThrow('scene graph is not valid XML');
    });
});

describe('nodeLabel', () => {
    it.each([
        ['Play', 'Button "Play" – Go'],
        ['Root', 'Scene "Root"'],
    ])('labels %s', (name, label) => {
        expect(nodeLabel(find(name))).toBe(label);
    });

    it('labels an element without name', () => {
        expect(nodeLabel(parsed().root.querySelector('Label') as Element)).toBe('Label');
    });
});

describe('xpathFor', () => {
    it('uses the name when there is one', () => {
        expect(xpathFor(find('Play'))).toBe("//Button[@Name='Play']");
    });

    it('escapes quotes in the name', () => {
        expect(xpathFor(find("O'Neil"))).toBe("//Button[@Name='O\\'Neil']");
    });

    it('uses the position below the root otherwise', () => {
        expect(xpathFor(parsed().root.querySelector('Label') as Element)).toBe('/Scene/Panel[2]/Label[1]');
    });

    it('is the root name for the root without name', () => {
        expect(xpathFor(parseSceneGraph('<Scene/>').root)).toBe('/Scene');
    });
});

describe('ancestorIds', () => {
    it('lists the ids from the root down', () => {
        const scene = parsed();
        const play = scene.byId.get('n2') as Element;
        expect(ancestorIds(scene, play)).toEqual(['n0', 'n1']);
        expect(ancestorIds(scene, scene.root)).toEqual([]);
    });

    it('skips elements the scene does not know', () => {
        const scene = parsed();
        const stranger = new DOMParser().parseFromString('<A><B/></A>', 'text/xml').querySelector('B') as Element;
        expect(ancestorIds(scene, stranger)).toEqual([]);
    });
});

describe('elementAt', () => {
    const root = (): Element => parsed().root;

    it('finds the deepest element at a point', () => {
        expect(elementAt(root(), {x: 20, y: 20})?.getAttribute('ID')).toBe('b1');
    });

    it('finds a container where no child is', () => {
        expect(elementAt(root(), {x: 190, y: 190})?.nodeName).toBe('Panel');
    });

    it('prefers the later sibling, which is drawn on top', () => {
        const overlap = new DOMParser().parseFromString(
            '<S RectangleX="10" RectangleY="10"><A ID="a" RectangleX="10" RectangleY="10"/>'
            + '<B ID="b" RectangleX="10" RectangleY="10"/></S>', 'text/xml').documentElement;
        expect(elementAt(overlap, {x: 1, y: 1})?.getAttribute('ID')).toBe('b');
    });

    it('finds nothing outside, and nothing for empty rectangles', () => {
        expect(elementAt(root(), {x: 500, y: 500})).toBeNull();
        expect(elementAt(parseSceneGraph('<S/>').root, {x: 0, y: 0})).toBeNull();
    });
});

describe('scenePointAt', () => {
    it('scales a position on the screen to scene units', () => {
        const rect = {left: 10, top: 20, width: 100, height: 200};
        expect(scenePointAt(rect, {width: 200, height: 400}, {x: 60, y: 120})).toEqual({x: 100, y: 200});
    });
});

describe('highlightStyle', () => {
    it('places the element in percent of the scene', () => {
        expect(highlightStyle(find('Play'), {width: 200, height: 400})).toEqual({left: '5%', top: '2.5%', width: '25%', height: '5%'});
    });

    it.each([
        ['nothing is selected', null, {width: 1, height: 1}],
        ['there is no scene', 'sel', null],
        ['the scene has no size', 'sel', {width: 0, height: 0}],
    ])('is null when %s', (_, selected, scene) => {
        expect(highlightStyle(selected ? find('Play') : null, scene)).toBeNull();
    });
});
