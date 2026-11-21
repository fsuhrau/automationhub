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
        ['anna', 'anna', false],
        ['token:ci', 'token ci', false],
        ['mcp:anna:sess1', 'agent of anna', true],
        ['mcp:token:ci:sess1', 'agent of token ci', true],
        ['mcp:anna', 'agent of anna', true],
    ])('%s is %s', (owner, label, agent) => {
        expect(describeController(owner)).toEqual({label, agent});
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

    it('shows a person without a lease owner line', async () => {
        render(<DeviceController owner="anna"/>);
        fireEvent.mouseOver(screen.getByLabelText('Controlled by anna'));
        expect(await screen.findByText('Controlled by anna')).toBeInTheDocument();
        expect(screen.queryByText(/Lease owner/)).toBeNull();
    });
});
