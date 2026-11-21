import {useState} from 'react';
import {DropzoneState, useDropzone} from 'react-dropzone';
import {useNow} from '../../hooks/useNow';
import {Loaded} from '../../hooks/useLoader';
import {IAppBinaryData} from '../../types/app';
import {Accept, acceptFor} from './bundlesModel';
import {BundleDialog} from './parts/BundleDialogs';
import {MenuTarget} from './parts/BundleMenu';
import {BundleActions, useBundleActions} from './useBundleActions';
import {useBundleList} from './useBundleList';
import {BundleScope, useBundleScope} from './useBundleScope';
import {BundleUploads, useBundleUpload} from './useBundleUpload';

const MINUTE = 60_000;

export interface BundlesPage {
    scope: BundleScope,
    now: number,
    loader: Loaded<IAppBinaryData[]>,
    accept: Accept,
    // the bundle that was uploaded last, to point it out
    fresh: number | null,
    uploads: BundleUploads,
    drop: DropzoneState,
    actions: BundleActions,
    menu: MenuTarget | null,
    setMenu: (target: MenuTarget | null) => void,
    dialog: BundleDialog | null,
    setDialog: (dialog: BundleDialog | null) => void,
}

// useBundlesPage is everything the list of bundles works with: the bundles,
// uploading by button or by dropping files, and the changes of a bundle
export const useBundlesPage = (): BundlesPage => {
    const scope = useBundleScope();
    const now = useNow(MINUTE);
    const loader = useBundleList(scope.ref);
    const accept = acceptFor(scope.app?.platform);
    const [fresh, setFresh] = useState<number | null>(null);
    const [menu, setMenu] = useState<MenuTarget | null>(null);
    const [dialog, setDialog] = useState<BundleDialog | null>(null);
    const uploads = useBundleUpload({
        app: scope.ref, accept, appName: scope.app?.name,
        onUploaded: bundle => {
            setFresh(bundle.id);
            loader.retry();
        },
    });
    const drop = useDropzone({onDrop: uploads.upload, noClick: true, noKeyboard: true, multiple: true});
    const actions = useBundleActions(scope.ref, loader.retry);
    return {scope, now, loader, accept, fresh, uploads, drop, actions, menu, setMenu, dialog, setDialog};
};
