import {describe, expect, it} from 'vitest';
import {act, fireEvent, screen} from '@testing-library/react';
import {PlatformType} from '../../types/platform.type.enum';
import {device} from './testKit';
import {open, withHistory} from './deviceShowKit';

describe('DeviceShowPage log and screen', () => {
    it('shows log lines as the stream sends them and its failure', async () => {
        const page = open(device({platformType: PlatformType.Android}), withHistory());
        await screen.findByRole('log', {name: 'Device log'});
        const source = page.log.sources[0];
        expect(source.url).toContain('/logs/stream');
        act(() => source.emit('lines', ['12-01 10:00:00.000  1  2 I Unity: started']));
        expect(await screen.findByText(/Unity: started/)).toBeInTheDocument();
        expect(screen.getByText('Live')).toBeInTheDocument();
        act(() => source.emit('failure', {message: 'adb is gone'}));
        expect(await screen.findAllByText('adb is gone')).not.toHaveLength(0);
        expect(screen.getByText('No log')).toBeInTheDocument();
    });

    it('shows no log for a device that is not an Android phone', async () => {
        open(device({platformType: PlatformType.iOS}), withHistory());
        await screen.findByText('Ready');
        expect(screen.queryByRole('log')).toBeNull();
    });

    it('takes a new screenshot and shows a missing one', async () => {
        open(device(), withHistory());
        const image = await screen.findByAltText('Screen of Pixel');
        fireEvent.click(screen.getByRole('button', {name: 'Take a new screenshot'}));
        fireEvent.error(screen.getByAltText('Screen of Pixel'));
        expect(await screen.findByText('The device returned no screenshot.')).toBeInTheDocument();
        expect(image).toHaveAttribute('src', expect.stringContaining('/screenshot'));
    });
});
