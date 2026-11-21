import React from 'react';
import {ConfirmDialog} from '../../../components/common';
import {IAppBinaryData} from '../../../types/app';
import {archiveText, deleteText} from '../bundleDialogsModel';
import {bundleVersion} from '../bundlesModel';

interface BundleDialogProps {
    bundle: IAppBinaryData,
    busy: boolean,
    onConfirm: () => void,
    onClose: () => void,
}

export const DeleteBundleDialog: React.FC<BundleDialogProps> = ({bundle, busy, onConfirm, onClose}) =>
    <ConfirmDialog open={true} title={`Delete version ${bundleVersion(bundle)}?`} confirm="Delete bundle" busy={busy}
        onConfirm={onConfirm} onClose={onClose}>
        {deleteText(bundle)}
    </ConfirmDialog>;

export const ArchiveBundleDialog: React.FC<BundleDialogProps> = ({bundle, busy, onConfirm, onClose}) =>
    <ConfirmDialog open={true} title={`Archive version ${bundleVersion(bundle)}?`} confirm="Archive bundle" busy={busy} color="warning"
        onConfirm={onConfirm} onClose={onClose}>
        {archiveText(bundle)}
    </ConfirmDialog>;
