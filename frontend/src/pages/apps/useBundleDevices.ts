import {useLoader} from '../../hooks/useLoader';
import {useDeviceService} from '../../services/device.service';
import IDeviceData from '../../types/device';
import {PlatformType} from '../../types/platform.type.enum';
import {platformsOf} from './bundleDetailModel';

// useBundleDevices are the devices of the project a bundle of this platform
// can go on; null while loading. Devices that can't be loaded count as none:
// the page is about the bundle.
export const useBundleDevices = (projectId: string, platform: PlatformType | undefined): IDeviceData[] | null => {
    const service = useDeviceService();
    const {data, error} = useLoader(async () => {
        const platforms = platformsOf(platform);
        if (platforms.length === 0) return [];
        const all = await service.list(projectId);
        return all.filter(d => platforms.includes(d.platformType));
    }, [service, projectId, platform]);
    return error ? [] : data;
};
