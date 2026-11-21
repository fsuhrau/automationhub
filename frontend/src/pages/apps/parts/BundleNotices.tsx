import React from 'react';
import {Alert, Button, LinearProgress} from '@mui/material';
import {ArchiveOutlined} from '@mui/icons-material';
import {formatDateTime} from '../../../components/DateTime';
import {IAppBundleDetail} from '../../../types/app';
import {isInstallable} from '../bundlesModel';

const ARCHIVED = 'The file is gone; the details and test runs stay for comparison. '
    + 'The bundle can\'t be installed or run anymore. Upload the same file again to bring it back.';

const unread = (archived: boolean): string => archived
    ? 'This bundle was uploaded before the master read all details from the file, and its file is archived.'
    : 'This bundle was uploaded before the master read all details from the file. Read it again to see them.';

interface BundleNoticesProps {
    detail: IAppBundleDetail,
    reading: boolean,
    onRead: () => void,
}

const Unread: React.FC<BundleNoticesProps> = ({detail, reading, onRead}) =>
    <Alert severity="info" sx={{mb: 2}} action={isInstallable(detail) &&
        <Button color="inherit" size="small" disabled={reading} onClick={onRead}>Read now</Button>}>
        {unread(!!detail.archivedAt)}
    </Alert>;

// BundleNotices say what is special about a bundle: archived, details not
// read, or a warning from the master
export const BundleNotices: React.FC<BundleNoticesProps> = props => {
    const {detail, reading} = props;
    return <>
        {reading && <LinearProgress sx={{mb: 2}}/>}
        {detail.archivedAt && <Alert severity="info" icon={<ArchiveOutlined/>} sx={{mb: 2}}>
            Archived {formatDateTime(detail.archivedAt)}. {ARCHIVED}
        </Alert>}
        {!detail.info && <Unread {...props}/>}
        {detail.info?.warnings?.map(w => <Alert key={w} severity="warning" sx={{mb: 2}}>{w}</Alert>)}
    </>;
};
