import {useAppService} from '../../services/app.service';
import {BundleRef} from '../../types/app.ref';
import {BundleCommand, useBundleCommand} from './useBundleCommand';

export interface DetailActions {
    // the bundle is being read again
    reading: boolean,
    // the bundle is being archived or deleted
    busy: boolean,
    // each resolves true when it worked; a failure shows in the error banner
    readAgain: () => Promise<boolean>,
    archive: () => Promise<boolean>,
    remove: () => Promise<boolean>,
}

interface Handlers {
    // the bundle changed on the master: load it again
    onChanged: () => void,
    onDeleted: () => void,
}

// useDetailActions changes the bundle the details page shows
export const useDetailActions = (ref: BundleRef | null, {onChanged, onDeleted}: Handlers): DetailActions => {
    const service = useAppService();
    const reader = useBundleCommand();
    const changer = useBundleCommand();
    const change = (command: BundleCommand, op: (target: BundleRef) => Promise<unknown>, done: () => void): Promise<boolean> =>
        ref ? command.run(async () => {
            await op(ref);
            done();
        }) : Promise.resolve(false);
    return {
        reading: reader.busy,
        busy: changer.busy,
        readAgain: () => change(reader, service.readBundle, onChanged),
        archive: () => change(changer, service.archiveBundle, onChanged),
        remove: () => change(changer, service.deleteBundle, onDeleted),
    };
};
