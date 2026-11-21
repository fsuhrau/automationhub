import {useAppService} from '../../services/app.service';
import {IAppBundleDetail} from '../../types/app';
import {AppRef} from '../../types/app.ref';
import {Loaded, useLoader} from '../../hooks/useLoader';

// useBundleDetail loads one bundle with what was read from its file; it
// loads nothing for a missing app or a bundle id that isn't a number
export const useBundleDetail = (app: AppRef | null, id: number): Loaded<IAppBundleDetail> => {
    const service = useAppService();
    return useLoader(
        () => app && !isNaN(id) ? service.bundle({...app, id}) : null,
        [service, app?.projectId, app?.appId, id],
    );
};
