import {afterEach, describe, expect, it, vi} from 'vitest';
import {fireEvent, screen, waitFor} from '@testing-library/react';
import {BASE, controlled, loaded, PageOptions, PageWorld, renderPage} from './control/pageKit';

afterEach(() => {
    vi.unstubAllGlobals();
});

const openInspect = async (options?: PageOptions): Promise<PageWorld> => {
    const page = renderPage(options);
    await loaded();
    fireEvent.click(screen.getByRole('tab', {name: 'Inspect'}));
    return page;
};

const loadScene = async (): Promise<void> => {
    fireEvent.click(screen.getByRole('button', {name: 'Load scene graph'}));
    await screen.findByText('Scene');
};

// the screen is shown at 100x200 css pixels, the scene is 200x400 units
const clickScreen = (x: number, y: number): void => {
    const canvas = screen.getByLabelText('Device screen').querySelector('canvas') as HTMLCanvasElement;
    canvas.getBoundingClientRect = () => ({left: 0, top: 0, width: 100, height: 200}) as DOMRect;
    fireEvent.pointerDown(canvas, {clientX: x, clientY: y, pointerId: 1});
};

const pickOnScreen = (x: number, y: number): void => {
    fireEvent.click(screen.getByRole('button', {name: 'Pick on screen'}));
    clickScreen(x, y);
};

describe('the control page: the scene graph', () => {
    it('loads the scene graph and selects an element of the tree', async () => {
        await openInspect();
        await loadScene();
        expect(screen.getByRole('button', {name: 'Reload'})).toBeInTheDocument();
        fireEvent.click(screen.getByText('Panel'));
        expect(await screen.findByText('/Scene/Panel[1]')).toBeInTheDocument();
        expect(screen.queryByTestId('highlight')).toBeInTheDocument();
    });

    it('picks the element under a click on the screen, highlights it and shows its XPath', async () => {
        await openInspect();
        await loadScene();
        pickOnScreen(15, 25);
        expect(await screen.findByText("//Button[@Name='Play']")).toBeInTheDocument();
        expect(screen.getByTestId('highlight')).toHaveStyle({left: '10%', top: '10%'});
        expect(screen.getByText('ID')).toBeInTheDocument();
    });

    it('switches to the Inspect tab when picking from another tab', async () => {
        renderPage();
        await loaded();
        fireEvent.click(screen.getByRole('tab', {name: 'Inspect'}));
        await loadScene();
        fireEvent.click(screen.getByRole('button', {name: 'Pick on screen'}));
        fireEvent.click(screen.getByRole('tab', {name: 'App'}));
        clickScreen(15, 25);
        expect(await screen.findByText("//Button[@Name='Play']")).toBeInTheDocument();
    });

    it('shows a missing app in the tab and leads to the App tab', async () => {
        await openInspect({answers: {[`GET ${BASE}/hierarchy`]: new Error('app not connected')}});
        fireEvent.click(screen.getByRole('button', {name: 'Load scene graph'}));
        expect(await screen.findByText('app not connected')).toBeInTheDocument();
        expect(screen.getAllByRole('alert')).toHaveLength(1);
        fireEvent.click(screen.getByRole('button', {name: 'App tab'}));
        expect(await screen.findByText(/Install, start or stop a bundle/)).toBeInTheDocument();
    });
});

describe('the control page: the selected element', () => {
    const selectButton = async (): Promise<void> => {
        await loadScene();
        pickOnScreen(15, 25);
        await screen.findByText("//Button[@Name='Play']");
    };

    it('taps the element under control', async () => {
        const page = await controlled();
        fireEvent.click(screen.getByRole('tab', {name: 'Inspect'}));
        await selectButton();
        fireEvent.click(screen.getByRole('button', {name: 'Tap it'}));
        await waitFor(() => expect(page.api.to('POST', `${BASE}/element/touch`).map(c => c.data)).toEqual([{elementId: 'b1'}]));
    });

    it('cannot tap without control', async () => {
        await openInspect();
        await selectButton();
        expect(screen.getByRole('button', {name: 'Tap it'})).toBeDisabled();
    });

    it('copies the XPath', async () => {
        const writeText = vi.fn(async () => undefined);
        vi.stubGlobal('navigator', {clipboard: {writeText}});
        await openInspect();
        await selectButton();
        fireEvent.click(screen.getByRole('button', {name: 'Copy XPath'}));
        await waitFor(() => expect(writeText).toHaveBeenCalledWith("//Button[@Name='Play']"));
    });

    it('says when the clipboard is not available', async () => {
        vi.stubGlobal('navigator', {});
        await openInspect();
        await selectButton();
        fireEvent.click(screen.getByRole('button', {name: 'Copy XPath'}));
        expect(await screen.findByRole('alert')).toHaveTextContent('Could not copy to the clipboard');
    });
});
