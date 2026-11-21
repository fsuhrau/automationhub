import React from 'react';
import {IAppBinaryData} from '../../../types/app';
import {BundleActions} from '../useBundleActions';
import {ArchiveBundleDialog, DeleteBundleDialog} from './BundleConfirmDialogs';
import {TagsDialog} from './TagsDialog';

// the question a row asks before it changes a bundle
export type BundleDialog =
    | { kind: 'delete', bundle: IAppBinaryData }
    | { kind: 'archive', bundle: IAppBinaryData }
    | { kind: 'tags', bundle: IAppBinaryData, text: string };

interface BundleDialogsProps {
    dialog: BundleDialog | null,
    actions: BundleActions,
    onChange: (dialog: BundleDialog | null) => void,
}

// BundleDialogs are the dialogs of the list: delete, archive and edit tags
export const BundleDialogs: React.FC<BundleDialogsProps> = ({dialog, actions, onChange}) => {
    const close = (): void => onChange(null);
    const closeWhenDone = async (done: Promise<boolean>): Promise<void> => {
        if (await done) close();
    };
    switch (dialog?.kind) {
        case 'delete':
            return <DeleteBundleDialog bundle={dialog.bundle} busy={actions.busy} onClose={close}
                onConfirm={() => closeWhenDone(actions.remove(dialog.bundle))}/>;
        case 'archive':
            return <ArchiveBundleDialog bundle={dialog.bundle} busy={actions.busy} onClose={close}
                onConfirm={() => closeWhenDone(actions.archive(dialog.bundle))}/>;
        case 'tags':
            return <TagsDialog edit={dialog} busy={actions.busy} onClose={close} onChange={text => onChange({...dialog, text})}
                onSave={() => closeWhenDone(actions.saveTags(dialog.bundle, dialog.text))}/>;
        default:
            return null;
    }
};
