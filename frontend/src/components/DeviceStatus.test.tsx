import {describe, expect, it} from 'vitest';
import {fireEvent, render, screen} from '@testing-library/react';
import DeviceStatus from './DeviceStatus';
import DeviceController, {describeController} from './DeviceController';
import {DeviceStateType} from '../types/deviceStateType';
import {PlatformType} from '../types/platform.type.enum';

describe('DeviceStatus', () => {
    it('shows just the name of a state that needs no help', () => {
        render(<DeviceStatus state={DeviceStateType.Booted}/>);
        expect(screen.getByText('Booted')).toBeInTheDocument();
    });

    it('replaces the name by a label', () => {
        render(<DeviceStatus state={DeviceStateType.Booted} label="How to fix"/>);
        expect(screen.getByText('How to fix')).toBeInTheDocument();
    });

    it.each([
        {state: DeviceStateType.Unauthorized, platform: PlatformType.Android, hint: /USB debugging is not authorized/},
        {state: DeviceStateType.Unauthorized, platform: PlatformType.iOS, hint: /doesn't trust the computer yet/},
        {state: DeviceStateType.Unauthorized, platform: PlatformType.IOSSimulator, hint: /doesn't trust the computer yet/},
        {state: DeviceStateType.RemoteDisconnected, platform: undefined, hint: /node doesn't see this device/},
        {state: DeviceStateType.NodeDisconnected, platform: undefined, hint: /isn't connected to the master/},
        {state: DeviceStateType.Unknown, platform: undefined, hint: /isn't known yet/},
    ])('explains $state on hover', async ({state, platform, hint}) => {
        render(<DeviceStatus state={state} platform={platform}/>);
        fireEvent.mouseOver(screen.getByText(/\w/));
        expect(await screen.findByText(hint)).toBeInTheDocument();
    });
});

describe('describeController', () => {
    it.each([
        ['anna', {label: 'anna', agent: false}],
        ['anna (user:7)', {label: 'anna', agent: false, userId: 7}],
        ['Ada Lovelace (user:12)', {label: 'Ada Lovelace', agent: false, userId: 12}],
        ['user:7', {label: 'user:7', agent: false, userId: 7}],
        ['evil (user:1) (user:7)', {label: 'evil (user:1)', agent: false, userId: 7}],
        ['anna (user:x)', {label: 'anna (user:x)', agent: false}],
        ['anna (user:)', {label: 'anna (user:)', agent: false}],
        ['anna(user:7)', {label: 'anna(user:7)', agent: false}],
        ['user:', {label: 'user:', agent: false}],
        ['anonymous', {label: 'anonymous', agent: false}],
        ['token:ci', {label: 'token ci', agent: false}],
        ['mcp:anna:sess1', {label: 'agent of anna', agent: true}],
        ['mcp:anna (user:7):sess1', {label: 'agent of anna', agent: true, userId: 7}],
        ['mcp:user:7:sess1', {label: 'agent of user:7', agent: true, userId: 7}],
        ['mcp:token:ci:sess1', {label: 'agent of token ci', agent: true}],
        ['mcp:anna', {label: 'agent of anna', agent: true}],
    ])('%s is %j', (owner, info) => {
        expect(describeController(owner)).toMatchObject(info);
        expect(describeController(owner).userId).toBe((info as {userId?: number}).userId);
    });
});

describe('DeviceController', () => {
    it('shows nothing for a free device', () => {
        const {container} = render(<DeviceController/>);
        expect(container).toBeEmptyDOMElement();
    });

    it('shows who controls the device, with the lease owner on hover', async () => {
        render(<DeviceController owner="token:ci"/>);
        const icon = screen.getByLabelText('Controlled by token ci');
        fireEvent.mouseOver(icon);
        expect(await screen.findByText('Lease owner: token:ci')).toBeInTheDocument();
    });

    it('shows a signed-in user without the user id suffix', async () => {
        render(<DeviceController owner="anna (user:7)"/>);
        fireEvent.mouseOver(screen.getByLabelText('Controlled by anna'));
        expect(await screen.findByText('Controlled by anna')).toBeInTheDocument();
        expect(screen.queryByText(/Lease owner/)).toBeNull();
    });

    it('shows the lease owner of an agent acting for a user', async () => {
        render(<DeviceController owner="mcp:anna (user:7):s1"/>);
        fireEvent.mouseOver(screen.getByLabelText('Controlled by agent of anna'));
        expect(await screen.findByText('Lease owner: mcp:anna (user:7):s1')).toBeInTheDocument();
    });

    it('shows a person without a lease owner line', async () => {
        render(<DeviceController owner="anna"/>);
        fireEvent.mouseOver(screen.getByLabelText('Controlled by anna'));
        expect(await screen.findByText('Controlled by anna')).toBeInTheDocument();
        expect(screen.queryByText(/Lease owner/)).toBeNull();
    });
});
