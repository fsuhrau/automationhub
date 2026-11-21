import {useEffect} from 'react';
import {useError} from '../../../ErrorProvider';
import {errorMessage} from '../../../lib/errorMessage';
import {useLoader} from '../../../hooks/useLoader';
import {useDeviceService} from '../../../services/device.service';
import IDeviceData from '../../../types/device';
import {PlatformType} from '../../../types/platform.type.enum';

// useTestDevices loads the devices a test can be picked for (those of the
// app's platform); null while loading. A failure is shown as the page error.
export const useTestDevices = (projectId: string, platform?: PlatformType): IDeviceData[] | null => {
    const {setError} = useError();
    const service = useDeviceService();
    const {data, error} = useLoader(() => service.list(projectId, platform), [service, projectId, platform]);
    useEffect(() => {
        if (error) setError(errorMessage(error));
    }, [error, setError]);
    return data;
};
