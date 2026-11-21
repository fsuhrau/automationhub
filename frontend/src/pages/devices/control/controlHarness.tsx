// Renders the control page for tests: a fake ApiClient that answers by
// "METHOD url", the providers of the real route, and a fake video player.
import {fireEvent, screen, waitFor} from '@testing-library/react';
import {expect} from 'vitest';
import {Route, Routes} from 'react-router-dom';
import {fakeApi, FakeApi, Reply} from '../../../testing/fakeApi';
import {aProject} from '../../../testing/fixtures';
import {projectRoute, renderWithProviders} from '../../../testing/renderWithProviders';
import {DeviceLockStatus} from '../../../types/device.control.lease';
import DeviceControlPage from '../DeviceControlPage';
import {APPS, BUNDLES, DEVICE, fakeSupport, FakeSupport, FREE, leaseOf, heldBy} from './fakes';

export const BASE = '/p1/device/7';

export const XML = `<Scene RectangleX="200" RectangleY="400">
  <Panel X="0" Y="0" RectangleX="200" RectangleY="200"><Button ID="b1" Name="Play" LabelText="Go" X="20" Y="40" RectangleX="40" RectangleY="20"/></Panel>
</Scene>`;

export const LOG = ['11-06 13:51:05.123  1  2 I App: hello', '11-06 13:51:06.000  1  2 E App: boom'];

// the hub of one device that is free, until somebody takes it
const hubAnswers = (world: { lock: DeviceLockStatus }): Record<string, Reply> => ({
    [`GET ${BASE}`]: DEVICE,
    'GET /p1/apps': APPS,
    'GET /p1/app/1/bundles': BUNDLES,
    'GET /p1/app/2/bundles': BUNDLES,
    [`GET ${BASE}/lock`]: () => world.lock,
    [`POST ${BASE}/lock`]: () => {
        world.lock = heldBy();
        return leaseOf();
    },
    [`DELETE ${BASE}/lock`]: () => {
        world.lock = FREE;
        return null;
    },
    [`GET ${BASE}/hierarchy`]: XML,
    [`GET ${BASE}/logs`]: {lines: LOG},
});

// the inputs answer with an empty body
const INPUTS = ['tap', 'swipe', 'text', 'key', 'element/touch', 'app/install', 'app/start', 'app/stop'];

export interface PageOptions {
    answers?: Record<string, Reply>,
    lock?: DeviceLockStatus,
    support?: FakeSupport,
}

export interface PageWorld {
    api: FakeApi,
    support: FakeSupport,
    world: { lock: DeviceLockStatus },
}

export const renderPage = ({answers = {}, lock = FREE, support = fakeSupport()}: PageOptions = {}): PageWorld => {
    const world = {lock};
    const inputs = Object.fromEntries(INPUTS.map(i => [`POST ${BASE}/${i}`, null]));
    const api = fakeApi({...hubAnswers(world), ...inputs, ...answers});
    renderWithProviders(
        <Routes>{projectRoute(<Route path="device/:deviceId/control" element={<DeviceControlPage support={support}/>}/>)}</Routes>,
        {api, url: '/project/p1/device/7/control', hub: {projects: [aProject('p1', 'P')]}},
    );
    return {api, support, world};
};

export const loaded = async (): Promise<void> => {
    await screen.findByRole('heading', {name: 'Lab phone'});
};

// waits until the hub has given the lease to this browser
export const takeControl = async (page: PageWorld): Promise<void> => {
    fireEvent.click(screen.getAllByRole('button', {name: 'Take control'})[0]);
    await screen.findByText('You control this device');
    await waitFor(() => expect(page.world.lock.locked).toBe(true));
};

export const controlled = async (options?: PageOptions): Promise<PageWorld> => {
    const page = renderPage(options);
    await loaded();
    await takeControl(page);
    return page;
};
