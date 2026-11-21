import React from 'react';
import {AppRoutes} from '../../../lib/routes';
import {copyText} from '../../../lib/clipboard';
import {IAppBinaryData} from '../../../types/app';
import {BundleDialog} from './BundleDialogs';
import {BundleMenu, MenuTarget} from './BundleMenu';

interface RowMenuProps {
    // null keeps the menu closed
    target: MenuTarget | null,
    bundles: IAppBinaryData[],
    routes: AppRoutes,
    onDialog: (dialog: BundleDialog) => void,
    onClose: () => void,
}

// RowMenu is the menu of a row of the list, wired to the page
export const RowMenu: React.FC<RowMenuProps> = ({target, bundles, routes, onDialog, onClose}) => {
    if (!target) return null;
    const {bundle} = target;
    const ask = (kind: 'archive' | 'delete'): void => {
        onDialog({kind, bundle});
        onClose();
    };
    return <BundleMenu
        target={target} onClose={onClose} detailUrl={routes.bundle(bundle.id)}
        compareUrl={`${routes.compareBundles()}?b=${bundle.id}`} hasPrevious={bundles.some(b => b.id < bundle.id)}
        onCopyHash={() => void copyText(bundle.hash).then(onClose)}
        onArchive={() => ask('archive')} onDelete={() => ask('delete')}/>;
};
