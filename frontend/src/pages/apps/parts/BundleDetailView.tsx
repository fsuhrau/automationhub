import React, {useState} from 'react';
import {Box, Stack} from '@mui/material';
import {IAppBundleDetail} from '../../../types/app';
import IDeviceData from '../../../types/device';
import {iconUrlOf} from '../bundlesModel';
import {BundleScope} from '../useBundleScope';
import {DetailActions} from '../useDetailActions';
import {BundleActions} from './BundleActions';
import {ArchiveBundleDialog, DeleteBundleDialog} from './BundleConfirmDialogs';
import {BundleHeader} from './BundleHeader';
import {BundleNotices} from './BundleNotices';
import {BundleSections} from './BundleSections';

interface BundleDetailViewProps {
    detail: IAppBundleDetail,
    scope: BundleScope,
    devices: IDeviceData[] | null,
    actions: DetailActions,
    now: number,
}

type Asking = 'archive' | 'delete' | null;

// BundleDetailView is a loaded bundle: header, notices, sections and the
// questions before archiving or deleting it
export const BundleDetailView: React.FC<BundleDetailViewProps> = ({detail, scope, devices, actions, now}) => {
    const [asking, setAsking] = useState<Asking>(null);
    const close = (): void => setAsking(null);
    const confirm = (act: () => Promise<boolean>) => async (): Promise<void> => {
        if (await act()) close();
    };
    return <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}, mt: 2}}>
        <Stack direction={{xs: 'column', md: 'row'}} spacing={2} useFlexGap={true} alignItems={{md: 'center'}} sx={{mb: 2}}>
            <BundleHeader detail={detail} iconUrl={iconUrlOf(scope.ref, detail)} platform={scope.app?.platform}/>
            <BundleActions detail={detail} routes={scope.routes} reading={actions.reading} onRead={actions.readAgain}
                onArchive={() => setAsking('archive')} onDelete={() => setAsking('delete')}/>
        </Stack>
        <BundleNotices detail={detail} reading={actions.reading} onRead={actions.readAgain}/>
        <BundleSections detail={detail} scope={scope} devices={devices} now={now}/>
        {asking === 'archive' && <ArchiveBundleDialog bundle={detail} busy={actions.busy} onClose={close} onConfirm={confirm(actions.archive)}/>}
        {asking === 'delete' && <DeleteBundleDialog bundle={detail} busy={actions.busy} onClose={close} onConfirm={confirm(actions.remove)}/>}
    </Box>;
};
