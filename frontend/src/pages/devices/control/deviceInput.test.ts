import {describe, expect, it, vi} from 'vitest';
import {VIDEO_TEXT_MAX_BYTES} from '../../../services/device.video';
import {createDeviceInput, DeviceInput} from './deviceInput';
import {DeviceControlApi} from '../../../types/device.control.api';
import {Runner} from './useGuardedAction';
import {FakePlayer, fakeService} from './fakes';

interface Setup {
    service: DeviceControlApi,
    player: FakePlayer,
    run: Runner,
    input: DeviceInput,
}

const setup = (over: { lease?: string | null, live?: boolean } = {}): Setup => {
    const service = fakeService();
    const player = new FakePlayer('', document.createElement('canvas'), {} as never);
    const run: Runner = vi.fn(async (action: () => Promise<unknown>) => {
        await action();
    });
    const lease = over.lease === undefined ? 'tok' : over.lease;
    const input = createDeviceInput({lease, service, run, liveInput: () => over.live ? player : null});
    return {service, player, run, input};
};

describe('createDeviceInput.sendKey', () => {
    it('presses the key over the video socket when it takes input', () => {
        const {input, player, run} = setup({live: true});
        input.sendKey('home');
        expect(player.sent).toEqual([{input: {kind: 'key', action: 'press', key: 'home'}, lease: 'tok'}]);
        expect(run).not.toHaveBeenCalled();
    });

    it('uses the REST API otherwise', () => {
        const {input, service} = setup();
        input.sendKey('home');
        expect(service.key).toHaveBeenCalledWith('tok', 'home');
    });

    it('sends nothing without a lease', () => {
        const {input, service, run} = setup({lease: null});
        input.sendKey('home');
        expect(run).not.toHaveBeenCalled();
        expect(service.key).not.toHaveBeenCalled();
    });
});

describe('createDeviceInput.typeText', () => {
    it('types over the video socket and says when done', () => {
        const {input, player} = setup({live: true});
        const done = vi.fn();
        input.typeText('hi', done);
        expect(player.sent[0].input).toEqual({kind: 'text', text: 'hi'});
        expect(done).toHaveBeenCalled();
    });

    it('types without a done callback', () => {
        const {input, player} = setup({live: true});
        input.typeText('hi');
        expect(player.sent).toHaveLength(1);
    });

    it('sends text that is too long for the socket through the REST API', async () => {
        const {input, service, player} = setup({live: true});
        const done = vi.fn();
        input.typeText('x'.repeat(VIDEO_TEXT_MAX_BYTES + 1), done);
        await vi.waitFor(() => expect(done).toHaveBeenCalled());
        expect(service.text).toHaveBeenCalledWith('tok', 'x'.repeat(VIDEO_TEXT_MAX_BYTES + 1));
        expect(player.sent).toHaveLength(0);
    });

    it('uses the REST API without a live player, with or without done', async () => {
        const {input, service} = setup();
        input.typeText('a');
        await vi.waitFor(() => expect(service.text).toHaveBeenCalledWith('tok', 'a'));
    });

    it('sends nothing without a lease', () => {
        const {input, service, run} = setup({lease: null});
        input.typeText('a', vi.fn());
        expect(run).not.toHaveBeenCalled();
        expect(service.text).not.toHaveBeenCalled();
    });
});
