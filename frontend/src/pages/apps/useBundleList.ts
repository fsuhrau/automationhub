import {useAppService} from '../../services/app.service';
import {IAppBinaryData} from '../../types/app';
import {AppRef} from '../../types/app.ref';
import {Loaded, useLoader} from '../../hooks/useLoader';
import {newestFirst} from './bundlesModel';

// useBundleList loads the bundles of an app, newest first
export const useBundleList = (app: AppRef | null): Loaded<IAppBinaryData[]> => {
    const service = useAppService();
    return useLoader(
        () => app ? service.bundles(app).then(newestFirst) : null,
        [service, app?.projectId, app?.appId],
    );
};
