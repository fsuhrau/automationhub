import {describe, expect, it, vi} from 'vitest';
import {act, renderHook, RenderHookResult} from '@testing-library/react';
import {fakeService} from './fakes';
import {screenElement} from './pointerFakes';
import {Runner} from './useGuardedAction';
import {SceneControl, useSceneGraph} from './useSceneGraph';

const XML = `<Scene RectangleX="200" RectangleY="400">
  <Panel X="0" Y="0" RectangleX="200" RectangleY="200"><Button ID="b1" Name="Play" X="20" Y="40" RectangleX="40" RectangleY="20"/></Panel>
</Scene>`;

// runs actions at once and hands failures to the handler of the call
const run: Runner = async (action, onError) => {
    try {
        await action();
    } catch (e) {
        onError?.(e);
    }
};

type Setup = RenderHookResult<SceneControl, unknown>;

const setup = (hierarchy: () => Promise<string> = async () => XML): Setup =>
    renderHook(() => useSceneGraph(fakeService({hierarchy: vi.fn(hierarchy)}), run));

const loaded = async (): Promise<Setup> => {
    const hook = setup();
    await act(async () => hook.result.current.load());
    return hook;
};

describe('useSceneGraph', () => {
    it('loads the scene graph and opens the root', async () => {
        const {result} = await loaded();
        expect(result.current.scene?.items[0].id).toBe('n0');
        expect(result.current.expanded).toEqual(['n0']);
        expect(result.current.selected).toBeNull();
        expect(result.current.error).toBeNull();
    });

    it('shows a failure in the tab', async () => {
        const {result} = setup(async () => Promise.reject(new Error('app not connected')));
        await act(async () => result.current.load());
        expect(result.current.error).toBe('app not connected');
        expect(result.current.scene).toBeNull();
    });

    it('shows text that is not XML as a failure', async () => {
        const {result} = setup(async () => '<<<');
        await act(async () => result.current.load());
        expect(result.current.error).toBe('scene graph is not valid XML');
    });

    it('clears an old failure on reload', async () => {
        let fail = true;
        const {result} = setup(async () => fail ? Promise.reject(new Error('no')) : XML);
        await act(async () => result.current.load());
        fail = false;
        await act(async () => result.current.load());
        expect(result.current.error).toBeNull();
        expect(result.current.scene).not.toBeNull();
    });
});

describe('useSceneGraph selection', () => {
    it('selects an element, opens the tree down to it and highlights it', async () => {
        const {result} = await loaded();
        const button = result.current.scene?.byId.get('n2') as Element;
        act(() => result.current.select(button));
        expect(result.current.selected).toBe(button);
        expect(result.current.expanded).toEqual(['n0', 'n1']);
        expect(result.current.highlight).toEqual({left: '10%', top: '10%', width: '20%', height: '5%'});
    });

    it('clears the selection', async () => {
        const {result} = await loaded();
        act(() => result.current.select(result.current.scene?.root ?? null));
        act(() => result.current.select(null));
        expect(result.current.selected).toBeNull();
        expect(result.current.highlight).toBeNull();
    });

    it('selects without a scene but expands nothing', () => {
        const {result} = setup();
        const element = document.createElement('div');
        act(() => result.current.select(element));
        expect(result.current.selected).toBe(element);
        expect(result.current.expanded).toEqual([]);
    });

    it('tracks what the user opens', async () => {
        const {result} = await loaded();
        act(() => result.current.setExpanded(['n1']));
        expect(result.current.expanded).toEqual(['n1']);
    });
});

describe('useSceneGraph picking', () => {
    it('toggles the pick mode', () => {
        const {result} = setup();
        act(() => result.current.togglePicking());
        expect(result.current.picking).toBe(true);
        act(() => result.current.togglePicking());
        expect(result.current.picking).toBe(false);
    });

    it('picks the element under a point of the screen', async () => {
        const {result} = await loaded();
        const {el} = screenElement();
        // the screen element is 100x200 at (10,10); the scene is 200x400
        act(() => result.current.pickAt(el, {x: 10 + 20, y: 10 + 25}));
        expect(result.current.selected?.getAttribute('ID')).toBe('b1');
    });

    it('picks nothing without a scene or without a scene size', async () => {
        const {el} = screenElement();
        const none = setup();
        act(() => none.result.current.pickAt(el, {x: 20, y: 20}));
        expect(none.result.current.selected).toBeNull();

        const flat = setup(async () => '<Scene/>');
        await act(async () => flat.result.current.load());
        act(() => flat.result.current.pickAt(el, {x: 20, y: 20}));
        expect(flat.result.current.selected).toBeNull();
    });
});
