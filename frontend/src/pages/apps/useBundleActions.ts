import {useAppService} from '../../services/app.service';
import {IAppBinaryData} from '../../types/app';
import {AppRef} from '../../types/app.ref';
import {joinTags} from './bundlesModel';
import {useBundleCommand} from './useBundleCommand';

export interface BundleActions {
    busy: boolean,
    // each resolves true when it worked; a failure shows in the error banner
    remove: (bundle: IAppBinaryData) => Promise<boolean>,
    archive: (bundle: IAppBinaryData) => Promise<boolean>,
    // text is the tags as typed, separated by commas
    saveTags: (bundle: IAppBinaryData, text: string) => Promise<boolean>,
}

// useBundleActions changes the bundles of the list; onChanged runs after
// each change so the list can reload
export const useBundleActions = (app: AppRef | null, onChanged: () => void): BundleActions => {
    const service = useAppService();
    const {busy, run} = useBundleCommand();
    const change = (op: (target: AppRef) => Promise<unknown>): Promise<boolean> =>
        app ? run(async () => {
            await op(app);
            onChanged();
        }) : Promise.resolve(false);
    return {
        busy,
        remove: b => change(target => service.deleteBundle({...target, id: b.id})),
        archive: b => change(target => service.archiveBundle({...target, id: b.id})),
        saveTags: (b, text) => change(target => service.updateTags({...target, id: b.id}, joinTags(text.split(',')))),
    };
};
